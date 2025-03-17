/*
# 1. 需要获取IP信息和项目信息
# 2. Ci机器需要单独的标签
*/
package main

import (
	//"context"
	"encoding/json"
	"fmt"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"log"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

// Prometheus 目标结构
type PrometheusTarget struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func generatePrometheusConfig(ips []string, projectName, hostname string) (target PrometheusTarget) {
	labels := make(map[string]string, 10)

	labels["project"] = projectName
	labels["job"] = projectName
	labels["hostname"] = hostname

	target = PrometheusTarget{
		Targets: ips,
		Labels:  labels,
	}

	return
}

func saveConfig(targets []PrometheusTarget, projectName string) error {
	file, err := json.MarshalIndent(targets, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}

	if err := os.WriteFile("./filesd/"+projectName+"_sd.json", file, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

func main() {
	// OpenStack 认证信息
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: "http://10.6.2.500:5000/v3", // OpenStack 认证地址
		Username:         "admin",                   // 用户名
		Password:         "password",            // 密码
		DomainName:       "default",                 // 用户域
		TenantName:       "admin",                   // 项目名称
	}

	// 创建认证客户端
	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		log.Fatalf("认证失败: %v", err)
	}

	// 创建身份服务客户端
	identityClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		log.Fatalf("创建身份服务客户端失败: %v", err)
	}

	// 获取项目
	allProject, err := projects.List(identityClient, projects.ListOpts{}).AllPages()
	if err != nil {
		log.Fatalf("获取项目列表失败", err)
	}
	projectList, err := projects.ExtractProjects(allProject)

	projectMap := make(map[string]string, len(projectList))
	for _, project := range projectList {
		if project.Name != "service" || project.Name != "devlop" {
			projectMap[project.ID] = project.Name
		}

	}

	// 创建计算服务客户端
	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: "RegionOne", // 区域名称
	})
	if err != nil {
		log.Fatalf("创建计算客户端失败: %v", err)
	}

	// 获取实例列表
	allPages, err := servers.List(client, servers.ListOpts{AllTenants: true}).AllPages()
	if err != nil {
		log.Fatalf("获取实例列表失败: %v", err)
	}

	// 解析实例列表
	allServers, err := servers.ExtractServers(allPages)
	if err != nil {
		log.Fatalf("解析实例列表失败: %v", err)
	}

	//var ips []string
	// 打印实例信息
	result := make(map[string][]PrometheusTarget)
	for _, server := range allServers {
		if server.Status == "ACTIVE" {
			addrMap := server.Addresses
			ipadd := addrMap["external-network-10.15.0"]
			var ip = ""
			for _, addr := range ipadd.([]interface{}) {
				ip = addr.(map[string]interface{})["addr"].(string) + ":9100"
				break
			}
			//ips = append(ips, ip)
			fmt.Printf("实例名称: %v,IP地址: %v,项目名称: %v\n", server.Name, ip, projectMap[server.TenantID])
			projectname := projectMap[server.TenantID]
			render := generatePrometheusConfig([]string{ip}, projectname, server.Name)
			targets := result[projectname]
			targets = append(targets, render)
			result[projectname] = targets
			//ips = nil
		} else {
			continue
		}
	}
	for key, value := range result {
		if err := saveConfig(value, key); err != nil {
			panic(err)
		}
	}
	log.Println("Prometheus 配置生成完成")
}
