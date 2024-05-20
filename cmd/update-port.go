package cmd

import (
    "context"
    "fmt"
    "log"
    "strings"

    "github.com/charmbracelet/bubbles/list"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/network"
    "github.com/docker/docker/client"
    "github.com/docker/go-connections/nat"
    "github.com/dustinkirkland/golang-petname"
    "github.com/spf13/cobra"
)

var (
    newPorts         []string
    newImageName     string
    newContainerName string
    hostIP           string
)

var updatePortCmd = &cobra.Command{
    Use:   "update-port",
    Short: "Update the port mappings of a running Docker container",
    Run: func(cmd *cobra.Command, args []string) {
        containerID := fuzzySearchContainer()
        if containerID == "" {
            fmt.Println("No container selected. Exiting.")
            return
        }

        if newImageName == "" {
            newImageName = petname.Generate(2, "-")
        }

        if newContainerName == "" {
            newContainerName = petname.Generate(2, "-")
        }

        updatePort(containerID, newPorts, newImageName, newContainerName, hostIP)
    },
}

func init() {
    rootCmd.AddCommand(updatePortCmd)
    updatePortCmd.Flags().StringSliceVarP(&newPorts, "new-ports", "p", []string{}, "New ports to map to the container in the format 'hostPort:containerPort' (required)")
    updatePortCmd.Flags().StringVarP(&newImageName, "new-image-name", "i", "", "Name for the new image (default is a random pet name)")
    updatePortCmd.Flags().StringVarP(&newContainerName, "new-container-name", "n", "", "Name for the new container (default is a random pet name)")
    updatePortCmd.Flags().StringVarP(&hostIP, "host-ip", "h", "127.0.0.1", "Host IP address")

    updatePortCmd.MarkFlagRequired("new-ports")
}

func fuzzySearchContainer() string {
    ctx := context.Background()
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        log.Fatalf("Error creating Docker client: %v", err)
    }

    containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
    if err != nil {
        log.Fatalf("Error listing containers: %v", err)
    }

    items := make([]list.Item, len(containers))
    for i, container := range containers {
        items[i] = listItem{
            name:  container.Names[0],
            id:    container.ID[:12],
            ports: container.Ports,
        }
    }

    p := tea.NewProgram(initialModel(items))
    model, err := p.Run()
    if err != nil {
        log.Fatalf("Error running TUI: %v", err)
    }

    selectedContainer := model.(model).selectedID
    for _, container := range containers {
        if strings.HasPrefix(container.ID, selectedContainer) {
            return container.ID
        }
    }

    return ""
}

func updatePort(containerID string, newPorts []string, newImageName, newContainerName, hostIP string) {
    ctx := context.Background()
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        log.Fatalf("Error creating Docker client: %v", err)
    }

    // Stop the container
    if err := cli.ContainerStop(ctx, containerID, nil); err != nil {
        log.Fatalf("Error stopping container: %v", err)
    }

    // Commit the container
    commitResponse, err := cli.ContainerCommit(ctx, containerID, types.ContainerCommitOptions{
        Reference: newImageName,
    })
    if err != nil {
        log.Fatalf("Error committing container: %v", err)
    }
    newImageID := commitResponse.ID

    // Get the container configuration
    containerJSON, err := cli.ContainerInspect(ctx, containerID)
    if err != nil {
        log.Fatalf("Error inspecting container: %v", err)
    }

    // Preserve old port mappings and add new ones
    oldPortBindings := containerJSON.HostConfig.PortBindings
    oldExposedPorts := containerJSON.Config.ExposedPorts

    newPortBindings := nat.PortMap{}
    newExposedPorts := nat.PortSet{}

    for _, portMapping := range newPorts {
        ports := strings.Split(portMapping, ":")
        hostPort, containerPort := ports[0], ports[1]
        newPortBindings[nat.Port(containerPort+"/tcp")] = []nat.PortBinding{{HostIP: hostIP, HostPort: hostPort}}
        newExposedPorts[nat.Port(containerPort+"/tcp")] = struct{}{}
    }

    for port, bindings := range oldPortBindings {
        newPortBindings[port] = bindings
    }

    for port := range oldExposedPorts {
        newExposedPorts[port] = struct{}{}
    }

    // Remove the old container
    if err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true}); err != nil {
        log.Fatalf("Error removing container: %v", err)
    }

    // Start a new container with the new port mappings
    resp, err := cli.ContainerCreate(ctx, &container.Config{
        Image:        newImageID,
        ExposedPorts: newExposedPorts,
    }, &container.HostConfig{
        PortBindings: newPortBindings,
    }, &network.NetworkingConfig{}, nil, newContainerName)
    if err != nil {
        log.Fatalf("Error creating container: %v", err)
    }

    if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
        log.Fatalf("Error starting container: %v", err)
    }

    fmt.Println("Container started with new port mappings")
}

type listItem struct {
    name  string
    id    string
    ports []types.Port
}

func (i listItem) Title() string       { return i.name }
func (i listItem) Description() string { return i.id }
func (i listItem) FilterValue() string { return i.name }

type model struct {
    list       list.Model
    selectedID string
}

func initialModel(items []list.Item) model {
    l := list.New(items, list.NewDefaultDelegate(), 0, 0)
    l.Title = "Select a container"
    return model{list: l}
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "enter":
            selectedItem, ok := m.list.SelectedItem().(listItem)
            if ok {
                m.selectedID = selectedItem.id
            }
            return m, tea.Quit
        }
    }

    return m, cmd
}

func (m model) View() string {
    return m.list.View()
}
