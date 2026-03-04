package main

import (
	"errors"
	realos "os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/io"
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	fakes "code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type FakeOS struct {
	exitCalled                 int
	exitCalledWithCode         int
	mkdirCalled                int
	mkdirCalledWithPath        string
	mkdirCalledWithMode        realos.FileMode
	removeCalled               int
	removeCalledWithPath       string
	symlinkCalled              int
	symlinkCalledWithTarget    string
	symlinkCalledWithSource    string
	readdirCalled              int
	readdirCalledWithPath      string
	readfileCalled             int
	readfileCalledWithPath     string
	writefileCalled            int
	writefileCalledWithPath    string
	writefileCalledWithContent []byte
	writefileCalledWithMode    realos.FileMode
	removeShouldReturnError    error
	readdirShouldReturn        []realos.DirEntry
	readfileShouldReturn       []byte
}

func (os *FakeOS) Exit(code int) {
	os.exitCalled++
	os.exitCalledWithCode = code
}

func (os *FakeOS) Mkdir(path string, mode realos.FileMode) error {
	os.mkdirCalled++
	os.mkdirCalledWithPath = path
	os.mkdirCalledWithMode = mode
	return nil
}

func (os *FakeOS) Remove(path string) error {
	os.removeCalled++
	os.removeCalledWithPath = path
	return os.removeShouldReturnError
}

func (os *FakeOS) Symlink(target string, source string) error {
	os.symlinkCalled++
	os.symlinkCalledWithTarget = target
	os.symlinkCalledWithSource = source
	return nil
}

func (os *FakeOS) ReadDir(path string) ([]realos.DirEntry, error) {
	os.readdirCalled++
	os.readdirCalledWithPath = path
	return os.readdirShouldReturn, nil
}

func (os *FakeOS) ReadFile(path string) ([]byte, error) {
	os.readfileCalled++
	os.readfileCalledWithPath = path
	return os.readfileShouldReturn, nil
}

func (os *FakeOS) WriteFile(path string, content []byte, mode realos.FileMode) error {
	os.writefileCalled++
	os.writefileCalledWithPath = path
	os.writefileCalledWithContent = content
	os.writefileCalledWithMode = mode
	return nil
}

var _ = Describe("TargetsPlugin", func() {

	var fakeCliConnection *fakes.FakeCliConnection
	var targetsPlugin *TargetsPlugin
	var fakeOS FakeOS

	BeforeEach(func() {
		fakeOS = FakeOS{}
		os = &fakeOS
		fakeCliConnection = &fakes.FakeCliConnection{}
		targetsPlugin = newTargetsPlugin()
	})

	Describe("Command Syntax", func() {
		It("displays usage when targets called with too many arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"targets", "blah"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "targets"}))
		})

		It("displays usage when set-target called with too many arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"set-target", "blah", "blah"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "set-target", "[-f]", "NAME"}))
		})

		It("displays usage when set-target called with too few arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"set-target"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "set-target", "[-f]", "NAME"}))
		})

		It("displays usage when set-target called with unsupported option", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"set-target", "blah", "-k"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "set-target", "[-f]", "NAME"}))
		})

		It("displays usage when save-target called with too many arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"save-target", "blah", "blah"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "save-target", "[-f]", "[NAME]"}))
		})

		It("displays usage when save-target called with unsupported option", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"save-target", "blah", "-k"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "save-target", "[-f]", "[NAME]"}))
		})

		It("displays usage when delete-target called with too few arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"delete-target"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "delete-target", "NAME"}))
		})

		It("displays usage when delete-target called with too many arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"delete-target", "blah", "blah"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "delete-target", "NAME"}))
		})

		It("displays proper first time message", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"targets"})
			})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(output).To(ContainSubstrings([]string{"No targets have been saved"}))
			Expect(output).To(ContainSubstrings([]string{"cf", "save-target", "NAME"}))
		})
	})

	Describe("DeleteTargetCommand", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = realos.MkdirTemp("", "cf-targets-test-*")
			Expect(err).NotTo(HaveOccurred())
			targetsPlugin.targetsPath = tmpDir
			targetsPlugin.currentPath = filepath.Join(tmpDir, "current")
		})

		AfterEach(func() {
			realos.RemoveAll(tmpDir)
		})

		It("deletes an existing target", func() {
			targetFile := filepath.Join(tmpDir, "mytest"+targetsPlugin.suffix)
			err := realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"delete-target", "mytest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(fakeOS.removeCalledWithPath).To(Equal(targetFile))
			Expect(output).To(ContainSubstrings([]string{"Deleted target", "mytest"}))
		})

		It("exits with error when Remove fails", func() {
			targetFile := filepath.Join(tmpDir, "mytest"+targetsPlugin.suffix)
			err := realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			fakeOS.removeShouldReturnError = errors.New("permission denied")

			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"delete-target", "mytest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Error:", "permission denied"}))
		})
	})

	Describe("Configuration File Manipulation", func() {

		It("creates the proper target directory", func() {
			Expect(fakeOS.mkdirCalled).To(Equal(1))
			Expect(fakeOS.mkdirCalledWithPath).To(HaveSuffix("/.cf/targets"))
		})

		It("properly saves first target", func() {
			targetsPlugin.Run(fakeCliConnection, []string{"save-target", "first"})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(fakeOS.writefileCalled).To(Equal(1))
			Expect(fakeOS.writefileCalledWithPath).To(HaveSuffix("/.cf/targets/first.config.json"))
			Expect(fakeOS.symlinkCalled).To(Equal(1))
			Expect(fakeOS.symlinkCalledWithSource).To(HaveSuffix("/.cf/targets/current"))
			Expect(fakeOS.symlinkCalledWithTarget).To(HaveSuffix("/.cf/targets/first.config.json"))
		})

		It("properly saves second target", func() {
			targetsPlugin.Run(fakeCliConnection, []string{"save-target", "first"})
			targetsPlugin.Run(fakeCliConnection, []string{"save-target", "second"})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(fakeOS.writefileCalled).To(Equal(2))
			Expect(fakeOS.writefileCalledWithPath).To(HaveSuffix("/.cf/targets/second.config.json"))
			Expect(fakeOS.removeCalledWithPath).To(HaveSuffix("/.cf/targets/current"))
			Expect(fakeOS.symlinkCalled).To(Equal(2))
			Expect(fakeOS.symlinkCalledWithSource).To(HaveSuffix("/.cf/targets/current"))
			Expect(fakeOS.symlinkCalledWithTarget).To(HaveSuffix("/.cf/targets/second.config.json"))
		})
	})
})
