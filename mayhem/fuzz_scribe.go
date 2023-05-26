package fuzz_scribe

import (
    fuzz "github.com/AdaLogics/go-fuzz-headers"

    "github.com/grafana/scribe/args"
    "github.com/grafana/scribe/cmdutil"
    "github.com/grafana/scribe/jsonnet"
    "github.com/grafana/scribe/stringutil"
)

func mayhemit(data []byte) int {

    if len(data) > 2 {
        num := int(data[0])
        data = data[1:]
        fuzzConsumer := fuzz.NewConsumer(data)
        
        switch num {
            
            case 0:
                var testArgs []string
                repeat, _ := fuzzConsumer.GetInt()

                for i := 0; i < repeat; i++ {

                    temp, _ := fuzzConsumer.GetString()
                    testArgs = append(testArgs, temp)
                }

                args.ParseArguments(testArgs)
                return 0

            case 1:
                var commands cmdutil.CommandOpts
                fuzzConsumer.GenerateStruct(&commands)

                cmdutil.StepCommand(commands)
                return 0

            case 2:
                var commands cmdutil.PipelineCommandOpts
                fuzzConsumer.GenerateStruct(&commands)

                cmdutil.PipelineCommand(commands)
                return 0

            case 3:
                testPath, _ := fuzzConsumer.GetString()
                
                jsonnet.Lint(testPath)
                return 0

            case 4:
                testPath, _ := fuzzConsumer.GetString()
                
                jsonnet.Format(testPath)
                return 0

            case 5:
                testLength, _ := fuzzConsumer.GetInt()
                
                stringutil.Random(testLength)
                return 0

            case 6:
                testString, _ := fuzzConsumer.GetString()
                
                stringutil.Slugify(testString)
                return 0
        }
    }
    return 0
}

func Fuzz(data []byte) int {
    _ = mayhemit(data)
    return 0
}