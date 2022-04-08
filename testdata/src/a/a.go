package a

type s0 struct { // want ".low cohesion, lcom is 2.*"
}

func (s s0) method1() {}
func (s s0) method2() {}

// receiver as a field
type s1 string

func (s s1) method1() string {
	return string(s)
}
func (s s1) method2() string {
	return string(s)
}

// embed
type embeddee struct {
	a int
}

func (e embeddee) method1() {}

type embedder struct {
	embeddee
}

func (e embedder) method2() {}

// pointer and value receiver
type s2 struct {
	m int
}

func (a *s2) method1() int {
	return a.m
}
func (a s2) method2() int {
	return a.m
}

type s3 struct {
	m int
}

func (a s3) method1() int {
	return a.m
}
func (a s3) method2() int {
	return a.m
}
