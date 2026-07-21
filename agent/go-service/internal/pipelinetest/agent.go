package pipelinetest

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"path/filepath"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

type agentSession struct {
	client *maa.AgentClient
	cmd    *exec.Cmd
	done   chan error
}

func startAgent(config Config, resource *maa.Resource, diagnostics *[]string, stderr io.Writer) (*agentSession, error) {
	if config.AgentMode == AgentDisabled {
		*diagnostics = append(*diagnostics, "Agent disabled by --no-agent")
		return nil, nil
	}
	if !fileExists(config.AgentPath) {
		if config.AgentMode == AgentAuto {
			*diagnostics = append(*diagnostics, fmt.Sprintf("Agent not started; auto-detected executable does not exist: %s", config.AgentPath))
			return nil, nil
		}
		return nil, fmt.Errorf("executable does not exist: %s", config.AgentPath)
	}

	port, err := reserveTCPPort()
	if err != nil {
		return nil, fmt.Errorf("reserve local TCP port: %w", err)
	}
	client, err := maa.NewAgentClient(maa.WithTcpPort(port))
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	identifier, err := client.Identifier()
	if err != nil {
		client.Destroy()
		return nil, fmt.Errorf("read generated identifier: %w", err)
	}
	if err := client.BindResource(resource); err != nil {
		client.Destroy()
		return nil, fmt.Errorf("bind Resource: %w", err)
	}

	cmd := exec.Command(config.AgentPath, identifier)
	cmd.Dir = agentWorkingDirectory(config.AgentPath)
	cmd.Stdout = stderr
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		client.Destroy()
		return nil, fmt.Errorf("start %s: %w", config.AgentPath, err)
	}
	session := &agentSession{client: client, cmd: cmd, done: make(chan error, 1)}
	go func() { session.done <- cmd.Wait() }()

	deadline := time.Now().Add(config.AgentTimeout)
	connectTimeout := 250 * time.Millisecond
	if config.AgentTimeout < connectTimeout {
		connectTimeout = config.AgentTimeout
	}
	if err := client.SetTimeout(connectTimeout); err != nil {
		session.close()
		return nil, fmt.Errorf("set timeout: %w", err)
	}
	var connectErr error
	for time.Now().Before(deadline) {
		select {
		case processErr := <-session.done:
			session.done = nil
			session.close()
			return nil, fmt.Errorf("process exited before connection: %v", processErr)
		default:
		}
		connectErr = client.Connect()
		if connectErr == nil {
			if err := client.SetTimeout(config.AgentTimeout); err != nil {
				session.close()
				return nil, fmt.Errorf("set connected timeout: %w", err)
			}
			registrations, listErr := client.GetCustomRecognitionList()
			if listErr != nil {
				session.close()
				return nil, fmt.Errorf("read registrations: %w", listErr)
			}
			*diagnostics = append(*diagnostics, fmt.Sprintf("Agent connected with %d Custom Recognition registration(s)", len(registrations)))
			return session, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	session.close()
	return nil, fmt.Errorf("connect timed out after %s: %w", config.AgentTimeout, connectErr)
}

func reserveTCPPort() (uint16, error) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	address, ok := listener.Addr().(*net.TCPAddr)
	if !ok || address.Port < 1 || address.Port > 65535 {
		return 0, fmt.Errorf("unexpected listener address: %s", listener.Addr())
	}
	return uint16(address.Port), nil
}

func (s *agentSession) close() {
	if s == nil {
		return
	}
	if s.client != nil {
		if s.client.Connected() {
			_ = s.client.Disconnect()
		}
		s.client.Destroy()
		s.client = nil
	}
	if s.cmd == nil || s.cmd.Process == nil {
		return
	}
	if s.done != nil {
		select {
		case <-s.done:
			return
		case <-time.After(2 * time.Second):
		}
	}
	_ = s.cmd.Process.Kill()
	if s.done != nil {
		select {
		case <-s.done:
		case <-time.After(time.Second):
		}
	}
}

func agentWorkingDirectory(agentPath string) string {
	agentDir := filepath.Dir(agentPath)
	parent := filepath.Dir(agentDir)
	if isDir(filepath.Join(parent, "maafw")) {
		return parent
	}
	return agentDir
}
