package fileglob

import (
	"github.com/franela/goblin"
	"os"
	"path/filepath"
	"testing"
)

func TestGlob(t *testing.T) {
	pwd, _ := os.Getwd()
	dir := filepath.Join(pwd, "test")
	g := goblin.Goblin(t)

	g.Describe("file glob", func() {
		g.It("fglob", func() {
			ignoreDirs := []string{".bzr", ".hg", ".git", ".idea", ".directory"}
			filters := []string{"JPG", "JPEG"}
			res, err := Glob(dir, filters, false, ignoreDirs)
			g.Assert(len(res)).Equal(8)
			g.Assert(err == nil).IsTrue()
		})
	})

	g.Describe("file glob stream", func() {
		g.It("stream files", func() {
			ignoreDirs := []string{".bzr", ".hg", ".git", ".idea", ".directory"}
			filters := []string{"JPG", "JPEG"}
			res := make([]string, 0)
			stream := GlobStream(dir, filters, false, ignoreDirs, 25)
			g.Assert(stream != nil).IsTrue()

			for buf := range stream {
				res = append(res, buf...)
			}
			g.Assert(len(res)).Equal(8)
		})
	})

}
