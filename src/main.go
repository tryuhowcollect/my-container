package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) < 3 {
		panic("引数が足りません")
	}
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("引数が不正です")
	}
}

func run() {
	fmt.Printf("Running %v parent\n", os.Args[2:])
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running command: %v\n", err)
	}
}

func child() {
	fmt.Printf("Running %v child\n", os.Args[2:])
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	// ホスト名を "container" に設定
	if err := syscall.Sethostname([]byte("container")); err != nil {
		fmt.Printf("Error setting hostname: %v\n", err)
		return
	}

	// proc ファイルシステムをマウント
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		fmt.Printf("Error mounting proc: %v\n", err)
		return
	}

	// Vethペアを作成してIPアドレスを設定
	if err := setupVeth(); err != nil {
		fmt.Printf("Error setting up veth: %v\n", err)
		return
	}

	// コマンドを実行
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running command: %v\n", err)
	}

	// proc ファイルシステムをアンマウント
	if err := syscall.Unmount("/proc", 0); err != nil {
		fmt.Printf("Error unmounting proc: %v\n", err)
	}
}

func setupVeth() error {
	// Vethペアを作成
	if err := exec.Command("ip", "link", "add", "veth0-host", "type", "veth", "peer", "name", "veth0-ct").Run(); err != nil {
		return fmt.Errorf("failed to create veth pair: %v", err)
	}

	// VethペアにIPアドレスを設定
	if err := exec.Command("ip", "addr", "add", "10.10.10.10/24", "dev", "veth0-host").Run(); err != nil {
		return fmt.Errorf("failed to assign IP to veth0-host: %v", err)
	}
	if err := exec.Command("ip", "link", "set", "up", "dev", "veth0-host").Run(); err != nil {
		return fmt.Errorf("failed to bring up veth0-host: %v", err)
	}

	// コンテナ側のVethインターフェースを新しいネットワーク名前空間に移動
	if err := moveToNetNS("veth0-ct"); err != nil {
		return fmt.Errorf("failed to move veth0-ct to netns: %v", err)
	}

	// 新しいネットワーク名前空間内での設定
	if err := exec.Command("ip", "netns", "exec", "netns01", "ip", "addr", "add", "10.10.10.11/24", "dev", "veth0-ct").Run(); err != nil {
		return fmt.Errorf("failed to assign IP to veth0-ct in netns: %v", err)
	}
	if err := exec.Command("ip", "netns", "exec", "netns01", "ip", "link", "set", "up", "dev", "veth0-ct").Run(); err != nil {
		return fmt.Errorf("failed to bring up veth0-ct in netns: %v", err)
	}

	return nil
}

func moveToNetNS(ifName string) error {
	fmt.Printf("Moving %s to new netns\n", ifName)

	// 新しいネットワーク名前空間を作成
	if err := exec.Command("ip", "netns", "add", "netns01").Run(); err != nil {
		return fmt.Errorf("failed to create netns: %v", err)
	}

	// インターフェースを新しいネットワーク名前空間に移動
	cmd := exec.Command("ip", "link", "set", ifName, "netns", "netns01")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error moving %s to netns: %v: %s", ifName, err, string(output))
	}

	return nil
}
