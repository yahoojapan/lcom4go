# LCOM4go
LCOM4go is a tool to compute LCOM4, Lack of Cohesion of Methods metrics ver.4, for golang projects.

# Install
```
$ go install --ldflags "-s -w" --trimpath github.com/yahoojapan/lcom4go/cmd/lcom4@latest
```

# Usage
```
$  go vet -vettool=$(go env GOPATH)/bin/lcom4 ./...
...

$  go vet -vettool=$(go env GOPATH)/bin/lcom4 net/http
...
```

# LCOM4 definition
https://objectscriptquality.com/docs/metrics/lack-cohesion-methods-lcom4


# Examples

The lcom4 of `s0` is 1 because both `method1` and `method2` use `s0.m`.
```
type s0 struct {
	m int
}

func (a s0) method1() int {
	return a.m
}
func (a s0) method2() int {
	return -a.m
}
```


The lcom4 of `s1` is 2 because `method3` uses `a.n` which is not used by `method1` and `method2`.
```
type s1 struct {
	m int
	n int
}

func (a s1) method1() int {
	return a.m
}
func (a s1) method2() int {
	return -a.m
}
func (a s1) method3() int {
	return -a.n
}
```


# Running the tests
```
go test ./...
```


# License

This software is released under the MIT License, see the license file.
# References
* https://www.aivosto.com/project/help/pm-oo-cohesion.html#LCOM4
* https://kenchon.github.io/cohesive-code
* https://objectscriptquality.com/docs/metrics/lack-cohesion-methods-lcom4
* https://github.com/FujiHaruka/eslint-plugin-lcom
* https://metacpan.org/release/JOENIO/Analizo-1.20.3/source/lib/Analizo/Metric/LackOfCohesionOfMethods.pm
* https://github.com/potfur/lcom
* http://www.isys.uni-klu.ac.at/PDF/1995-0043-MHBM.pdf
* https://github.com/cleuton/jqana
