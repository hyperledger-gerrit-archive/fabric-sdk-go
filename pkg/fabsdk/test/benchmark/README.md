#Benchmark testing of Channel Client
    This benchmark test has 5 calls of Channel Client's Execute() function
    The first 4 are invalid and the last one is valid. They are all executed in each benchmark's iterations.
    Under the directory where this file resides, the test commands are run as shown under the below comments: 
	(
	    * on a Macbook Pro, warning messages are stripped out below for conciseness
	    * Benchmark is using Go's test command with -bench=. the period is for current folder
	    * the -run=notest flag means execute a non-existant 'notest' in the current folder
	        This will avoid running normal unit tests along with the benchmarks
	    * by default, the benchmark tool decides when it collected enough information and stops
	    * the use of -benchtime=XXs forces the benchmark to keep executing until this time has elapsed
	        This allows the tool to run for longer times and collect more accurate information for larger execution loads
	    * the benchmark output format is as follows:
	        benchmarkname           [logs from benchamark tests-They have removed from the example commands below]   NbOfIterationExecutions     TimeInNanoSeconds/Iteration Executed
	        Example from below commands:
	        BenchmarkExecuteTx-8    [logs removed]                                                                   100000                      164854 ns/op
	        
	    * the command output also shows the environment and the package used for the benchmark exection:
	        goos: darwin
            goarch: amd64
            pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
	)

TODO Need a more controlled benchmark about channel client (perhaps do perf profiling to get more fine grained memory/performance issues)

$ go test -run=notest -bench=.
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	       1	1051368879 ns/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	1.826s
$ go test -run=notest -bench=. -benchtime=10s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	   100000	    164854 ns/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	18.054s
$ go test -run=notest -bench=. -benchtime=30s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	   300000	    185493 ns/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	58.387s
$ go test -run=notest -bench=. -benchtime=60s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	   500000	    232575 ns/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	120.207s
$ go test -run=notest -bench=. -benchtime=120s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	  1000000	    360515 ns/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	365.571s