package container

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/grussorusso/serverledge/internal/config"
	//	"github.com/docker/docker/pkg/stdcopy"
)

type DockerFactory struct {
	cli *client.Client
	ctx context.Context
}

func InitDockerContainerFactory() *DockerFactory {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	dockerFact := &DockerFactory{cli, ctx}
	cf = dockerFact
	return dockerFact
}

func (cf *DockerFactory) Create(image string, opts *ContainerOptions) (ContainerID, error) {
	log.Println("IMAGE NAME: ", image)
	if !cf.HasImage(image) {
		log.Printf("Pulling image: %s", image)
		pullResp, err := cf.cli.ImagePull(cf.ctx, image, types.ImagePullOptions{})
		if err != nil {
			log.Printf("Could not pull image: %s", image)
			// we do not return here, as a stale copy of the image
			// could still be available locally
		} else {
			defer pullResp.Close()
			// This seems to be necessary to wait for the image to be pulled:
			io.Copy(ioutil.Discard, pullResp)
			log.Printf("Pulled image: %s", image)
			refreshedImages[image] = true
		}
	}

	var networkConfig *network.NetworkingConfig

	if image != "marypie/serverledge-python-redis:latest" {
		networkConfig = nil
	} else {
		log.Println("OK")
		//network config for Cloud node connection to Redis DB
		networkConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{},
		}
		networkConfig.EndpointsConfig["redis"] = &network.EndpointSettings{
			NetworkID: "redis-net",
		}
	}

	portCont := nat.PortSet{
		"8080/tcp": struct{}{},
	}
	portHost := nat.PortMap{
		"8080/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "8080",
			},
		},
	}
	log.Printf("PORTA CONTAINER: %s", portCont)
	log.Printf("PORTA HOST: %s", portHost)
	resp, err := cf.cli.ContainerCreate(cf.ctx, &container.Config{
		Image:        image,
		ExposedPorts: portCont, //----added
		Cmd:          opts.Cmd,
		Env:          opts.Env,
		Tty:          false,
	}, &container.HostConfig{
		Resources:    container.Resources{Memory: opts.MemoryMB * 1048576}, // convert to bytes
		PortBindings: portHost,                                             //----added
	}, networkConfig, nil, "")

	id := resp.ID

	r, err := cf.cli.ContainerInspect(cf.ctx, id)
	log.Printf("Container %s has name %s", id, r.Name)

	return id, err
}

func (cf *DockerFactory) CopyToContainer(contID ContainerID, content io.Reader, destPath string) error {
	return cf.cli.CopyToContainer(cf.ctx, contID, destPath, content, types.CopyToContainerOptions{})
}

func (cf *DockerFactory) Start(contID ContainerID) error {
	if err := cf.cli.ContainerStart(cf.ctx, contID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	return nil
}

func (cf *DockerFactory) Destroy(contID ContainerID) error {
	// force set to true causes running container to be killed (and then
	// removed)
	return cf.cli.ContainerRemove(cf.ctx, contID, types.ContainerRemoveOptions{Force: true})
}

func (cf *DockerFactory) HasImage(image string) bool {
	// TODO: we should try using cf.cli.ImageList(...)
	cmd := fmt.Sprintf("docker images %s | grep -vF REPOSITORY", image)
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return false
	}

	// We have the image, but we may need to refresh it
	if config.GetBool(config.FACTORY_REFRESH_IMAGES, false) {
		if refreshed, ok := refreshedImages[image]; !ok || !refreshed {
			return false
		}
	}
	return true
}

func (cf *DockerFactory) GetIPAddress(contID ContainerID) (string, error) {
	contJson, err := cf.cli.ContainerInspect(cf.ctx, contID)
	if err != nil {
		return "", err
	}
	/*log.Println("conJSON.NetworkSettings: ", contJson.NetworkSettings)
	log.Println("contJson.NetworkSettings.Networks['redis'].IPAddress: ", contJson.NetworkSettings.Networks["redis"].IPAddress)
	log.Println("contJson.NetworkSettings.IPAddress: ", contJson.NetworkSettings.IPAddress)*/
	return contJson.NetworkSettings.IPAddress, nil
}

func (cf *DockerFactory) GetMemoryMB(contID ContainerID) (int64, error) {
	contJson, err := cf.cli.ContainerInspect(cf.ctx, contID)
	if err != nil {
		return -1, err
	}
	return contJson.HostConfig.Memory / 1048576, nil
}
