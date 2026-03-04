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
	readlineCalled             int
	readlineShouldReturn       string
	readlineShouldReturnError  error
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

func (os *FakeOS) ReadLine() (string, error) {
	os.readlineCalled++
	return os.readlineShouldReturn, os.readlineShouldReturnError
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

	Describe("checkStatus bug fix", func() {
		It("sets currentNeedsSaving to false when no symlink exists", func() {
			tmpDir, err := realos.MkdirTemp("", "cf-targets-test-*")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = realos.RemoveAll(tmpDir) }()

			targetsPlugin.currentPath = filepath.Join(tmpDir, "nonexistent-current")
			targetsPlugin.checkStatus()
			Expect(targetsPlugin.status.currentNeedsSaving).To(BeFalse())
			Expect(targetsPlugin.status.currentHasName).To(BeFalse())
		})

		It("allows set-target without -f when no symlink exists", func() {
			var tmpDir string
			var err error
			tmpDir, err = realos.MkdirTemp("", "cf-targets-test-*")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = realos.RemoveAll(tmpDir) }()

			targetsPlugin.targetsPath = tmpDir
			targetsPlugin.currentPath = filepath.Join(tmpDir, "current")

			targetFile := filepath.Join(tmpDir, "mytest"+targetsPlugin.suffix)
			err = realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"set-target", "mytest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(output).To(ContainSubstrings([]string{"Set target to", "mytest"}))
		})
	})

	Describe("SwitchTargetCommand", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = realos.MkdirTemp("", "cf-targets-test-*")
			Expect(err).NotTo(HaveOccurred())
			targetsPlugin.targetsPath = tmpDir
			targetsPlugin.currentPath = filepath.Join(tmpDir, "current")
		})

		AfterEach(func() {
			_ = realos.RemoveAll(tmpDir)
		})

		It("displays usage when called with no arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"switch-target"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "switch-target"}))
		})

		It("errors when target does not exist", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"switch-target", "nonexistent"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"does not exist"}))
		})

		It("skips save with -f and just switches", func() {
			targetFile := filepath.Join(tmpDir, "dest"+targetsPlugin.suffix)
			err := realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"switch-target", "-f", "dest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(fakeOS.writefileCalled).To(Equal(1)) // config only
			Expect(output).To(ContainSubstrings([]string{"Set target to", "dest"}))
		})

		It("auto-saves named current target before switching", func() {
			targetFile := filepath.Join(tmpDir, "dest"+targetsPlugin.suffix)
			err := realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			// Simulate named current target with unsaved changes
			targetsPlugin.status = TargetStatus{true, "origin", true, false}

			output := CaptureOutput(func() {
				targetsPlugin.SwitchTargetCommand([]string{"switch-target", "dest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(fakeOS.writefileCalled).To(Equal(2)) // save + switch
			Expect(output).To(ContainSubstrings([]string{"Saved current target as", "origin"}))
			Expect(output).To(ContainSubstrings([]string{"Set target to", "dest"}))
		})

		It("just switches when named current has no changes", func() {
			targetFile := filepath.Join(tmpDir, "dest"+targetsPlugin.suffix)
			err := realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			targetsPlugin.status = TargetStatus{true, "origin", false, false}

			output := CaptureOutput(func() {
				targetsPlugin.SwitchTargetCommand([]string{"switch-target", "dest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(fakeOS.writefileCalled).To(Equal(1)) // config only
			Expect(output).To(ContainSubstrings([]string{"Set target to", "dest"}))
		})

		It("saves unnamed target with --save-as before switching", func() {
			targetFile := filepath.Join(tmpDir, "dest"+targetsPlugin.suffix)
			err := realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			targetsPlugin.status = TargetStatus{false, "", true, false}

			output := CaptureOutput(func() {
				targetsPlugin.SwitchTargetCommand([]string{"switch-target", "--save-as", "dev", "dest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(fakeOS.writefileCalled).To(Equal(2))
			Expect(output).To(ContainSubstrings([]string{"Saved current target as", "dev"}))
			Expect(output).To(ContainSubstrings([]string{"Set target to", "dest"}))
		})

		It("prompts interactively for unnamed target and saves with entered name", func() {
			targetFile := filepath.Join(tmpDir, "dest"+targetsPlugin.suffix)
			err := realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			targetsPlugin.status = TargetStatus{false, "", true, false}
			fakeOS.readlineShouldReturn = "myname"

			output := CaptureOutput(func() {
				targetsPlugin.SwitchTargetCommand([]string{"switch-target", "dest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(fakeOS.readlineCalled).To(Equal(1))
			Expect(fakeOS.writefileCalled).To(Equal(2))
			Expect(output).To(ContainSubstrings([]string{"Saved current target as", "myname"}))
			Expect(output).To(ContainSubstrings([]string{"Set target to", "dest"}))
		})

		It("errors when interactive prompt returns empty name", func() {
			targetFile := filepath.Join(tmpDir, "dest"+targetsPlugin.suffix)
			err := realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			targetsPlugin.status = TargetStatus{false, "", true, false}
			fakeOS.readlineShouldReturn = ""

			output := CaptureOutput(func() {
				defer func() {
					if r := recover(); r != nil {
						if code, ok := r.(int); ok {
							fakeOS.Exit(code)
						} else {
							panic(r)
						}
					}
				}()
				targetsPlugin.SwitchTargetCommand([]string{"switch-target", "dest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"No name provided"}))
			Expect(output).To(ContainSubstrings([]string{"-f"}))
			Expect(output).To(ContainSubstrings([]string{"--save-as"}))
		})

		It("errors when ReadLine returns an error", func() {
			targetFile := filepath.Join(tmpDir, "dest"+targetsPlugin.suffix)
			err := realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			targetsPlugin.status = TargetStatus{false, "", true, false}
			fakeOS.readlineShouldReturnError = errors.New("input closed")

			output := CaptureOutput(func() {
				defer func() {
					if r := recover(); r != nil {
						if code, ok := r.(int); ok {
							fakeOS.Exit(code)
						} else {
							panic(r)
						}
					}
				}()
				targetsPlugin.SwitchTargetCommand([]string{"switch-target", "dest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Error:", "input closed"}))
		})

		It("just switches on fresh install (no symlink, currentNeedsSaving=false)", func() {
			targetFile := filepath.Join(tmpDir, "dest"+targetsPlugin.suffix)
			err := realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			// Fresh install: checkStatus sets currentNeedsSaving=false (after bug fix)
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"switch-target", "dest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(fakeOS.readlineCalled).To(Equal(0))
			Expect(fakeOS.writefileCalled).To(Equal(1)) // config only
			Expect(output).To(ContainSubstrings([]string{"Set target to", "dest"}))
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
