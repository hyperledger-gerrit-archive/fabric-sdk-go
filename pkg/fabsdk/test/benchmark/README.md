#Benchmark testing of Channel Client
    This benchmark test has 1 valid call of Channel Client's Execute() function
    
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
	        benchmarkname           [logs from benchamark tests-They have removed from the example commands below]   NbOfOperationExecutions     TimeInNanoSeconds/OperationExecuted   MemoryAllocated/OperationExecuted    NbOfAllocations/OperationExecuted  
	        Example from below commands:
	        BenchmarkExecuteTx-8    [logs removed]                                                                   100000                      164854 ns/op                          5743056 B/op                         50449 allocs/op 
	        
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
BenchmarkExecuteTx-8   	       1	1090034480 ns/op	 5743056 B/op	   50449 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	1.751s
$ go test -run=notest -bench=. -benchtime=10s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	   500000	     24705 ns/op	    5435 B/op	     102 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	16.731s
$ go test -run=notest -bench=. -benchtime=30s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	   2000000	     24700 ns/op	    5412 B/op	     102 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	63.496s
$ go test -run=notest -bench=. -benchtime=60s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	   3000000	     25051 ns/op	    5410 B/op	     102 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	103.539s
$ go test -run=notest -bench=. -benchtime=120s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	  10000000	     24858 ns/op	    5408 B/op	     102 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	275.085s