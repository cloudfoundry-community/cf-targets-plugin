package main

import (
	"encoding/json"
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
	readfileShouldReturnMap    map[string][]byte
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
	if os.readfileShouldReturnMap != nil {
		if data, ok := os.readfileShouldReturnMap[path]; ok {
			return data, nil
		}
	}
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

		It("auto-saves named current target and shows diff before switching", func() {
			targetFile := filepath.Join(tmpDir, "dest"+targetsPlugin.suffix)
			err := realos.WriteFile(targetFile, []byte("{}"), 0600)
			Expect(err).NotTo(HaveOccurred())

			// Provide differing JSON for showDiff: currentPath (saved) vs configPath (live)
			savedJSON, _ := json.MarshalIndent(map[string]interface{}{
				"AccessToken": "tok", "RefreshToken": "ref", "UAAOAuthClientSecret": "sec",
				"ColorEnabled": "true",
			}, "", " ")
			liveJSON, _ := json.MarshalIndent(map[string]interface{}{
				"AccessToken": "tok", "RefreshToken": "ref", "UAAOAuthClientSecret": "sec",
				"ColorEnabled": "false",
			}, "", " ")
			fakeOS.readfileShouldReturnMap = map[string][]byte{
				targetsPlugin.currentPath: savedJSON,
				targetsPlugin.configPath:  liveJSON,
			}

			// Simulate named current target with unsaved changes
			targetsPlugin.status = TargetStatus{true, "origin", true, false}

			output := CaptureOutput(func() {
				targetsPlugin.SwitchTargetCommand([]string{"switch-target", "dest"})
			})
			Expect(fakeOS.exitCalled).To(Equal(0))
			Expect(fakeOS.writefileCalled).To(Equal(2)) // save + switch
			// Diff should appear before the save message
			Expect(output).To(ContainSubstrings([]string{"---", "Current"}))
			Expect(output).To(ContainSubstrings([]string{"+++", "Target"}))
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

	Describe("createBuildMeta", func() {
		It("returns os.arch when no build metadata", func() {
			Expect(createBuildMeta("darwin", "arm64", "")).To(Equal("darwin.arm64"))
		})

		It("returns os.arch.build when build metadata provided", func() {
			Expect(createBuildMeta("linux", "amd64", "ci")).To(Equal("linux.amd64.ci"))
		})

		It("trims whitespace from all parts", func() {
			Expect(createBuildMeta("  darwin ", " arm64  ", " ci ")).To(Equal("darwin.arm64.ci"))
		})

		It("panics when os is empty", func() {
			Expect(func() { createBuildMeta("", "arm64", "") }).To(PanicWith(ContainSubstring("Go meta data is missing")))
		})

		It("panics when arch is empty", func() {
			Expect(func() { createBuildMeta("darwin", "", "") }).To(PanicWith(ContainSubstring("Go meta data is missing")))
		})

		It("panics when os is only whitespace", func() {
			Expect(func() { createBuildMeta("   ", "arm64", "") }).To(PanicWith(ContainSubstring("Go meta data is missing")))
		})
	})

	Describe("createSemVer", func() {
		It("returns major.minor.patch", func() {
			Expect(createSemVer("1", "2", "3", "", "")).To(Equal("1.2.3"))
		})

		It("appends prerelease with hyphen", func() {
			Expect(createSemVer("1", "2", "3", "beta", "")).To(Equal("1.2.3-beta"))
		})

		It("appends build with plus", func() {
			Expect(createSemVer("1", "2", "3", "", "linux.amd64")).To(Equal("1.2.3+linux.amd64"))
		})

		It("appends both prerelease and build", func() {
			Expect(createSemVer("1", "2", "3", "dev", "darwin.arm64")).To(Equal("1.2.3-dev+darwin.arm64"))
		})

		It("trims whitespace from all parts", func() {
			Expect(createSemVer(" 1 ", " 2 ", " 3 ", " rc1 ", " meta ")).To(Equal("1.2.3-rc1+meta"))
		})

		It("panics when major is empty", func() {
			Expect(func() { createSemVer("", "2", "3", "", "") }).To(PanicWith(ContainSubstring("Semanic version is missing")))
		})

		It("panics when minor is empty", func() {
			Expect(func() { createSemVer("1", "", "3", "", "") }).To(PanicWith(ContainSubstring("Semanic version is missing")))
		})

		It("panics when patch is empty", func() {
			Expect(func() { createSemVer("1", "2", "", "", "") }).To(PanicWith(ContainSubstring("Semanic version is missing")))
		})
	})

	Describe("showDiff", func() {
		makeJSON := func(overrides map[string]interface{}) []byte {
			base := map[string]interface{}{
				"AccessToken":          "token-abc",
				"RefreshToken":         "refresh-xyz",
				"UAAOAuthClientSecret": "secret-123",
				"Target":               "https://api.example.com",
				"ColorEnabled":         "true",
			}
			for k, v := range overrides {
				base[k] = v
			}
			data, err := json.MarshalIndent(base, "", " ")
			Expect(err).NotTo(HaveOccurred())
			return data
		}

		It("prints unified diff when files differ", func() {
			currentJSON := makeJSON(map[string]interface{}{"ColorEnabled": "true"})
			targetJSON := makeJSON(map[string]interface{}{"ColorEnabled": "false"})

			fakeOS.readfileShouldReturnMap = map[string][]byte{
				targetsPlugin.currentPath:             currentJSON,
				targetsPlugin.targetPath("other"): targetJSON,
			}

			output := CaptureOutput(func() {
				targetsPlugin.showDiff(targetsPlugin.targetPath("other"))
			})
			Expect(output).To(ContainSubstrings([]string{"---", "Current"}))
			Expect(output).To(ContainSubstrings([]string{"+++", "Target"}))
			Expect(output).To(ContainSubstrings([]string{`"true"`}))
			Expect(output).To(ContainSubstrings([]string{`"false"`}))
		})

		It("prints no differences when files are identical", func() {
			jsonData := makeJSON(nil)

			fakeOS.readfileShouldReturnMap = map[string][]byte{
				targetsPlugin.currentPath:             jsonData,
				targetsPlugin.targetPath("same"): jsonData,
			}

			output := CaptureOutput(func() {
				targetsPlugin.showDiff(targetsPlugin.targetPath("same"))
			})
			Expect(output).To(ContainSubstrings([]string{"hmmm no differences"}))
		})

		It("redacts sensitive fields in diff output", func() {
			currentJSON := makeJSON(map[string]interface{}{
				"AccessToken":  "current-token",
				"RefreshToken": "current-refresh",
			})
			targetJSON := makeJSON(map[string]interface{}{
				"AccessToken":  "different-token",
				"RefreshToken": "different-refresh",
			})

			fakeOS.readfileShouldReturnMap = map[string][]byte{
				targetsPlugin.currentPath:                currentJSON,
				targetsPlugin.targetPath("redacted"): targetJSON,
			}

			output := CaptureOutput(func() {
				targetsPlugin.showDiff(targetsPlugin.targetPath("redacted"))
			})
			// Tokens should be redacted
			Expect(output).To(ContainSubstrings([]string{"REDACTED sha256("}))
			// Raw tokens should NOT appear
			for _, line := range output {
				Expect(line).NotTo(ContainSubstring("current-token"))
				Expect(line).NotTo(ContainSubstring("different-token"))
				Expect(line).NotTo(ContainSubstring("current-refresh"))
				Expect(line).NotTo(ContainSubstring("different-refresh"))
			}
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
