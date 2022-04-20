package drone_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	shipwright "github.com/grafana/shipwright"
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/testutil"
	"github.com/sirupsen/logrus"
)

// testDemoPipeline tests a pipeline located in "demo" folder. the "path" argument should be relative to the demo folder in the root of the project.
// This function will do a basic equivalency check on what is generated by running the pipeline with the drone mode and what is in the "gen_drone.yml" file in the provided folder.
// Some standard arguments will be provided, like "-mode=drone", '-build-id="test"', "-path={path}", -log-level="debug".
func testDemoPipeline(t *testing.T, path string) {
	t.Helper()

	var (
		buf          = bytes.NewBuffer(nil)
		stderr       = bytes.NewBuffer(nil)
		ctx          = context.Background()
		pipelinePath = filepath.Join("../../../../demo", path)
	)

	testutil.RunPipeline(ctx, t, pipelinePath, io.MultiWriter(buf, os.Stdout), stderr, &plumbing.PipelineArgs{
		BuildID:  "test",
		Mode:     plumbing.RunModeDrone,
		Path:     fmt.Sprintf("./demo/%s", path), // Note that we're intentionally using ./demo/ instead of filepath because this path is used in a Go command.
		LogLevel: logrus.DebugLevel,
	})

	t.Log(stderr.String())

	expected, err := os.Open(filepath.Join(pipelinePath, "gen_drone.yml"))
	if err != nil {
		t.Fatal(err)
	}

	testutil.ReadersEqual(t, buf, expected)
}

func TestDroneClient(t *testing.T) {
	t.Run("It should generate a simple Drone pipeline",
		testutil.WithTimeout(time.Second*10, func(t *testing.T) {
			testDemoPipeline(t, "basic")
		}),
	)
	t.Run("It should generate a more complex multi Drone pipeline",
		testutil.WithTimeout(time.Second*10, func(t *testing.T) {
			testDemoPipeline(t, "multi")
		}),
	)
	t.Run("It should generate a multi-drone pipeline with a sub-pipeline",
		testutil.WithTimeout(time.Second*10, func(t *testing.T) {
			testDemoPipeline(t, "multi-sub")
		}),
	)
}

func TestDroneRun(t *testing.T) {
	t.Run("It should run sequential steps sequentially",
		testutil.WithTimeout(time.Second*5, func(t *testing.T) {
			t.SkipNow()

			t.Log("Creating new drone client...")
			sw := testutil.NewShipwright(shipwright.NewDroneClient)

			t.Log("Creating new test steps...")
			var (
				step1Chan = make(chan bool)
				step1     = testutil.NewTestStep(step1Chan)

				step2Chan = make(chan bool)
				step2     = testutil.NewTestStep(step2Chan)

				step3Chan = make(chan bool)
				step3     = testutil.NewTestStep(step3Chan)
			)

			t.Log("Running steps...")
			sw.Run(step1, step2, step3)

			go func() {
				t.Log("Done()")
				sw.Done()
				t.Log("done with Done()")
			}()

			var (
				expectedOrder = []int{1, 2, 3}
				order         = []int{}
			)

			t.Log("Waiting for order...")
			// Only watch for 3 channels
			for i := 0; i < 3; i++ {
				select {
				case <-step1Chan:
					order = append(order, 1)
				case <-step2Chan:
					order = append(order, 2)
				case <-step3Chan:
					order = append(order, 3)
				}
			}

			if !cmp.Equal(order, expectedOrder) {
				t.Fatal("Steps ran in unexpected order:", cmp.Diff(order, expectedOrder))
			}
		}))
}
