FROM golang:1.18 as builder

COPY . /scribe
WORKDIR /scribe/mayhem

RUN go install github.com/dvyukov/go-fuzz/go-fuzz@latest github.com/dvyukov/go-fuzz/go-fuzz-build@latest
RUN go get github.com/dvyukov/go-fuzz/go-fuzz-dep
RUN go get github.com/AdaLogics/go-fuzz-headers
RUN apt update && apt install -y clang

RUN go-fuzz-build -libfuzzer -o fuzz_scribe.a && \
    clang -fsanitize=fuzzer fuzz_scribe.a -o fuzz_scribe.libfuzzer

FROM debian:bookworm-slim
COPY --from=builder /scribe/mayhem/fuzz_scribe.libfuzzer /