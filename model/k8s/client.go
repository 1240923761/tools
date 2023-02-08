package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
	common "tools/common/util"
)

var (
	Client *K8SClient
)

type K8SClient struct {
	Client *kubernetes.Clientset
	Config K8SConfig
}
type K8SConfig struct {
	Host  string
	Token string
	Port  int
}

func MustInitK8SClient() {
	var err error
	Client = new(K8SClient)
	Client.Config = K8SConfig{
		Host:  "10.220.10.43",
		Port:  6443,
		Token: "eyJhbGciOiJSUzI1NiIsImtpZCI6IjV0eEN2TkkyeGhuRUs4UVl5ZGN3UVdWdjNFUm9DdnRuYmFTZ3F0eVRadkUifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJrdWJlLXN5c3RlbSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJ0ei1hZG1pbi10b2tlbi14dmNxZyIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJ0ei1hZG1pbiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjZlNGJlMWRlLTkxMTYtNDNhZi1hYjA2LWMyMGYzMjg3MjgwYyIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDprdWJlLXN5c3RlbTp0ei1hZG1pbiJ9.fYpHvh3xFA3ucsI1Jl7-U16WouEsgXTCAEdRGNSYdiE2E3UW5mKkawFplsh5gsJeEPWgbf6DOCJUcpZikK0pLRAM9FBK1kJHoci0rZgdu9jrNmXUmdSpwianMwDcNqge5er7-sstmn1rd_2yDR-fa-kElE4SiIabucVHuJxu9GAB8_UgS7euugi0dtkoJBh-L4n1iUGOitadCX3LKPTZZnfgE1BkMNp-09zK66pcRxUTNNddxqLN_YbqwOLEa8y2qgrm_JJICYhFi4QErnPMYmqM3r8qTCOyms9B6HULgu0_lNHDPAY4oT5UaaKS4SngdAkecvuMQ7kcV_dOMr9Mkg",
	}
	kubeConf := &rest.Config{
		Host:        fmt.Sprintf("%s:%d", Client.Config.Host, Client.Config.Port),
		BearerToken: Client.Config.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	//fmt.Println(kubeConf.BearerToken)
	Client.Client, err = kubernetes.NewForConfig(kubeConf)
	if err != nil {
		fmt.Println("init k8s err:", err.Error())
		panic("k8s err")
	}
	fmt.Println("init k8s success")
}

// GetAllNamespace get all namespace in cluster.

func (c *K8SClient) GetAllNamespace() []string {
	var namespaces []string
	namespaceList, err := c.Client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("get namespaces err:", err.Error())
	} else {
		//fmt.Printf(namespaces[0])
		for _, nsList := range namespaceList.Items {
			namespaces = append(namespaces, nsList.Name)
		}
	}

	return namespaces
}
func (c *K8SClient) GetNodes() {
	var nodes *v1.NodeList
	nodes, err := c.Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("get nodes err: ", err.Error())
	} else {
		//fmt.Printf(namespaces[0])
	}
	nodeNames := make([]string, 0)
	for i := 0; i < len(nodes.Items); i++ {
		//bs, _ := json.MarshalIndent(nodes.Items[i], " ", "\t")
		//fmt.Println(string(bs))
		nodeNames = append(nodeNames, nodes.Items[i].Name)
	}
	bs, _ := json.MarshalIndent(nodeNames, " ", "\t")
	fmt.Println(string(bs))
}

type Job struct {
	Kind string `yaml:"kind"`
}

func SplitYaml(file *[]byte, files *[][]byte, jobs *[]*Job) {
	var err error
	strs := strings.Split(common.Bytes2String(*file), "---")

	for i := 0; i < len(strs); i++ {
		job := new(Job)
		err = yaml.Unmarshal(common.String2Bytes(strs[i]), job)
		if err != nil {
			fmt.Println("unmarshal file err", err.Error())
		}
		*jobs = append(*jobs, job)
		*files = append(*files, common.String2Bytes(strs[i]))
	}
}
func ExecuteService(file *[]byte) error {
	service := new(v1.Service)
	err := yaml.Unmarshal(*file, service)

	if err != nil {
		return err
	}
	bs, _ := json.MarshalIndent(service, "", "\t")
	fmt.Println(common.Bytes2String(bs))
	return nil
}
func ExecuteDeployment(file *[]byte) error {
	deploy := new(appsv1.Deployment)
	err := yaml.Unmarshal(*file, deploy)
	if err != nil {
		return err
	}
	bs, _ := json.MarshalIndent(deploy, "", "\t")
	fmt.Println(common.Bytes2String(bs))
	return nil
}
func (c *K8SClient) ParseYaml(r io.Reader) {
	var bs = make([]byte, 1024)
	bf := bytes.NewBuffer(bs)
	io.Copy(bf, r)
	var (
		jobs  = make([]*Job, 0, 8)
		files = make([][]byte, 0, 8)
	)
	bs = bf.Bytes()
	SplitYaml(&bs, &files, &jobs)
	for i := 0; i < len(jobs); i++ {
		switch strings.ToLower(jobs[i].Kind) {
		case "service":
			fmt.Println("service")
			ExecuteService(&files[i])
		case "deployment":
			fmt.Println("deploy")
			ExecuteDeployment(&files[i])
		}
	}
}
