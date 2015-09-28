package discover

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/Sirupsen/logrus"
)

func BenchmarkReference(b *testing.B) {
	log := logrus.New()
	r, w := io.Pipe()
	go io.Copy(ioutil.Discard, r)
	log.Out = (w)
	for i := 0; i < b.N; i++ {
		log.Println("hello")
	}
}

func BenchmarkWrap(b *testing.B) {
	type wrap struct{ *logrus.Logger }
	log := &wrap{logrus.New()}
	r, w := io.Pipe()
	go io.Copy(ioutil.Discard, r)
	log.Out = (w)
	for i := 0; i < b.N; i++ {
		log.Println("hello")
	}
}

func BenchmarkLogger(b *testing.B) {
	log := NewLogger()
	r, w := io.Pipe()
	go io.Copy(ioutil.Discard, r)
	log.SetOutput(w)
	for i := 0; i < b.N; i++ {
		log.Println("hello")
	}
}
