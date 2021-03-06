package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// docker          run <container> <cmd> <args>
// go run main.go  run             <cmd> <args>
func main() {

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("unknown command")
	}
}

func run() {
	fmt.Printf("running parent pid:%d %v\n", os.Getpid(), os.Args[2:])
	// cmd := exec.Command("/proc/self/exe", os.Args[2], os.Args[3:]...)
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	errout(cmd.Run())
}

func child() {
	fmt.Printf("running child pid:%d %v\n", os.Getpid(), os.Args[2:])
	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cg()
	errout(syscall.Chroot("/home/jaked/os/ubuntu"))
	errout(syscall.Mount("proc", "/proc", "proc", 0, ""))

	errout(cmd.Run())

	errout(syscall.Unmount("/proc", 0))
}

func cg() {
	cgroups := "/sys/fs/cgroup/"

	pids := filepath.Join(cgroups, "pids")
	os.Mkdir(filepath.Join(pids, "brownbagX"), 0755)
	errout(ioutil.WriteFile(filepath.Join(pids, "brownbagX/pids.max"), []byte("5"), 0700))
	// Removes the new cgroup in place after the container exits
	errout(ioutil.WriteFile(filepath.Join(pids, "brownbagX/notify_on_release"), []byte("1"), 0700))
	errout(ioutil.WriteFile(filepath.Join(pids, "brownbagX/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func errout(err error) {
	if err != nil {
		panic(err.Error())
	}
}
