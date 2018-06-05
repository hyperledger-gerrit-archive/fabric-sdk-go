#Benchmark testing of Channel Client
    This benchmark test has 1 valid call of Channel Client's Execute() function
    
    Under the directory where this file resides, the test commands are run as shown under the below comments: 
	(
	    * on a Macbook Pro, warning messages are stripped out below for conciseness
	    * Benchmark is using Go's test command with -bench=ExecuteTx
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

$ go test -run=notest -bench=ExecuteTx
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	   50000	     24977 ns/op	    6049 B/op	     107 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	3.461s
$ go test -run=notest -bench=ExecuteTx -benchtime=10s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	   500000	     25605 ns/op	    6030 B/op	     107 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	15.384s
$ go test -run=notest -bench=ExecuteTx -benchtime=30s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	   2000000	     26316 ns/op	    6008 B/op	     107 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	80.548s
$ go test -run=notest -bench=ExecuteTx -benchtime=60s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	   3000000	     26066 ns/op	    6008 B/op	     107 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	106.784s
$ go test -run=notest -bench=ExecuteTx -benchtime=120s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
BenchmarkExecuteTx-8   	  10000000	     26240 ns/op	    6008 B/op	     107 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark	290.989s