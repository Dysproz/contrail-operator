package v1alpha1

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	configtemplates "github.com/Juniper/contrail-operator/pkg/apis/contrail/v1alpha1/templates"
	"github.com/Juniper/contrail-operator/pkg/certificates"
)

// +kubebuilder:validation:Enum=noauth;keystone
type AuthenticationMode string

const (
	AuthenticationModeNoAuth   AuthenticationMode = "noauth"
	AuthenticationModeKeystone AuthenticationMode = "keystone"
)

// +kubebuilder:validation:Enum=noauth;rbac
type AAAMode string

const (
	AAAModeNoAuth AAAMode = "no-auth"
	AAAModeRBAC   AAAMode = "rbac"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Config is the Schema for the configs API.
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=configs,scope=Namespaced
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.status.replicas`
// +kubebuilder:printcolumn:name="Ready_Replicas",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.endpoint`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Active",type=boolean,JSONPath=`.status.active`
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec   `json:"spec,omitempty"`
	Status ConfigStatus `json:"status,omitempty"`
}

// ConfigSpec is the Spec for the Config API.
// +k8s:openapi-gen=true
type ConfigSpec struct {
	CommonConfiguration  PodConfiguration    `json:"commonConfiguration,omitempty"`
	ServiceConfiguration ConfigConfiguration `json:"serviceConfiguration"`
}

// ConfigConfiguration is the Spec for the Config API.
// +k8s:openapi-gen=true
type ConfigConfiguration struct {
	Containers                  []*Container       `json:"containers,omitempty"`
	APIPort                     *int               `json:"apiPort,omitempty"`
	AnalyticsPort               *int               `json:"analyticsPort,omitempty"`
	CollectorPort               *int               `json:"collectorPort,omitempty"`
	RedisPort                   *int               `json:"redisPort,omitempty"`
	ApiIntrospectPort           *int               `json:"apiIntrospectPort,omitempty"`
	SchemaIntrospectPort        *int               `json:"schemaIntrospectPort,omitempty"`
	DeviceManagerIntrospectPort *int               `json:"deviceManagerIntrospectPort,omitempty"`
	SvcMonitorIntrospectPort    *int               `json:"svcMonitorIntrospectPort,omitempty"`
	AnalyticsApiIntrospectPort  *int               `json:"analyticsMonitorIntrospectPort,omitempty"`
	CollectorIntrospectPort     *int               `json:"collectorMonitorIntrospectPort,omitempty"`
	CassandraInstance           string             `json:"cassandraInstance,omitempty"`
	ZookeeperInstance           string             `json:"zookeeperInstance,omitempty"`
	NodeManager                 *bool              `json:"nodeManager,omitempty"`
	RabbitmqUser                string             `json:"rabbitmqUser,omitempty"`
	RabbitmqPassword            string             `json:"rabbitmqPassword,omitempty"`
	RabbitmqVhost               string             `json:"rabbitmqVhost,omitempty"`
	LogLevel                    string             `json:"logLevel,omitempty"`
	KeystoneSecretName          string             `json:"keystoneSecretName,omitempty"`
	KeystoneInstance            string             `json:"keystoneInstance,omitempty"`
	AuthMode                    AuthenticationMode `json:"authMode,omitempty"`
	AAAMode                     AAAMode            `json:"aaaMode,omitempty"`
	Storage                     Storage            `json:"storage,omitempty"`
	FabricMgmtIP                string             `json:"fabricMgmtIP,omitempty"`
	// Time (in hours) that the analytics object and log data stays in the Cassandra database. Defaults to 48 hours.
	AnalyticsDataTTL *int `json:"analyticsDataTTL,omitempty"`
	// Time (in hours) the analytics config data entering the collector stays in the Cassandra database. Defaults to 2160 hours.
	AnalyticsConfigAuditTTL *int `json:"analyticsConfigAuditTTL,omitempty"`
	// Time to live (TTL) for statistics data in hours. Defaults to 4 hours.
	AnalyticsStatisticsTTL *int `json:"analyticsStatisticsTTL,omitempty"`
	// Time to live (TTL) for flow data in hours. Defaults to 2 hours.
	AnalyticsFlowTTL *int `json:"analyticsFlowTTL,omitempty"`
}

// +k8s:openapi-gen=true
type ConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Active        *bool                             `json:"active,omitempty"`
	Nodes         map[string]string                 `json:"nodes,omitempty"`
	Ports         ConfigStatusPorts                 `json:"ports,omitempty"`
	ConfigChanged *bool                             `json:"configChanged,omitempty"`
	ServiceStatus map[string]ConfigServiceStatusMap `json:"serviceStatus,omitempty"`
	Endpoint      string                            `json:"endpoint,omitempty"`
}

type ConfigServiceStatusMap map[string]ConfigServiceStatus

type ConfigConnectionInfo struct {
	Name          string   `json:"name,omitempty"`
	Status        string   `json:"status,omitempty"`
	ServerAddress []string `json:"serverAddress,omitempty"`
}

type ConfigServiceStatus struct {
	NodeName    string `json:"nodeName,omitempty"`
	ModuleName  string `json:"moduleName,omitempty"`
	ModuleState string `json:"state"`
	Description string `json:"description,omitempty"`
}

type ConfigStatusPorts struct {
	APIPort       string `json:"apiPort,omitempty"`
	AnalyticsPort string `json:"analyticsPort,omitempty"`
	CollectorPort string `json:"collectorPort,omitempty"`
	RedisPort     string `json:"redisPort,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ConfigList contains a list of Config.
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}

const DMRunModeFull = "Full"

func (c *Config) InstanceConfiguration(request reconcile.Request,
	podList *corev1.PodList,
	client client.Client) error {
	instanceConfigMapName := request.Name + "-" + "config" + "-configmap"
	configMapInstanceDynamicConfig := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: instanceConfigMapName, Namespace: request.Namespace}, configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}

	cassandraNodesInformation, err := NewCassandraClusterConfiguration(c.Spec.ServiceConfiguration.CassandraInstance,
		request.Namespace, client)
	if err != nil {
		return err
	}

	zookeeperNodesInformation, err := NewZookeeperClusterConfiguration(c.Spec.ServiceConfiguration.ZookeeperInstance,
		request.Namespace, client)
	if err != nil {
		return err
	}

	rabbitmqNodesInformation, err := NewRabbitmqClusterConfiguration(c.Labels["contrail_cluster"],
		request.Namespace, client)
	if err != nil {
		return err
	}
	var rabbitmqSecretUser string
	var rabbitmqSecretPassword string
	var rabbitmqSecretVhost string
	if rabbitmqNodesInformation.Secret != "" {
		rabbitmqSecret := &corev1.Secret{}
		err = client.Get(context.TODO(), types.NamespacedName{Name: rabbitmqNodesInformation.Secret, Namespace: request.Namespace}, rabbitmqSecret)
		if err != nil {
			return err
		}
		rabbitmqSecretUser = string(rabbitmqSecret.Data["user"])
		rabbitmqSecretPassword = string(rabbitmqSecret.Data["password"])
		rabbitmqSecretVhost = string(rabbitmqSecret.Data["vhost"])
	}

	configConfigInterface := c.ConfigurationParameters()
	configConfig := configConfigInterface.(ConfigConfiguration)
	if rabbitmqSecretUser == "" {
		rabbitmqSecretUser = configConfig.RabbitmqUser
	}
	if rabbitmqSecretPassword == "" {
		rabbitmqSecretPassword = configConfig.RabbitmqPassword
	}
	if rabbitmqSecretVhost == "" {
		rabbitmqSecretVhost = configConfig.RabbitmqVhost
	}
	var collectorServerList, analyticsServerList, apiServerList, analyticsServerSpaceSeparatedList,
		apiServerSpaceSeparatedList, redisServerSpaceSeparatedList string
	var podIPList []string
	for _, pod := range podList.Items {
		podIPList = append(podIPList, pod.Status.PodIP)
	}
	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })
	sort.SliceStable(podIPList, func(i, j int) bool { return podIPList[i] < podIPList[j] })

	collectorServerList = strings.Join(podIPList, ":"+strconv.Itoa(*configConfig.CollectorPort)+" ")
	collectorServerList = collectorServerList + ":" + strconv.Itoa(*configConfig.CollectorPort)
	analyticsServerList = strings.Join(podIPList, ",")
	apiServerList = strings.Join(podIPList, ",")
	analyticsServerSpaceSeparatedList = strings.Join(podIPList, ":"+strconv.Itoa(*configConfig.AnalyticsPort)+" ")
	analyticsServerSpaceSeparatedList = analyticsServerSpaceSeparatedList + ":" + strconv.Itoa(*configConfig.AnalyticsPort)
	apiServerSpaceSeparatedList = strings.Join(podIPList, ":"+strconv.Itoa(*configConfig.APIPort)+" ")
	apiServerSpaceSeparatedList = apiServerSpaceSeparatedList + ":" + strconv.Itoa(*configConfig.APIPort)
	redisServerSpaceSeparatedList = strings.Join(podIPList, ":"+strconv.Itoa(*configConfig.RedisPort)+" ")
	redisServerSpaceSeparatedList = redisServerSpaceSeparatedList + ":" + strconv.Itoa(*configConfig.RedisPort)

	var data = make(map[string]string)
	for idx, pod := range podList.Items {
		configAuth, err := c.AuthParameters(client)
		if err != nil {
			return err
		}
		configIntrospectNodes := make([]string, 0)
		introspectPorts := map[string]int{
			"contrail-api":            *configConfig.ApiIntrospectPort,
			"contrail-schema":         *configConfig.SchemaIntrospectPort,
			"contrail-device-manager": *configConfig.DeviceManagerIntrospectPort,
			"contrail-svc-monitor":    *configConfig.SvcMonitorIntrospectPort,
			"contrail-analytics-api":  *configConfig.AnalyticsApiIntrospectPort,
			"contrail-collector":      *configConfig.CollectorIntrospectPort,
		}
		for service, port := range introspectPorts {
			nodesPortStr := pod.Status.PodIP + ":" + strconv.Itoa(port) + "::" + service
			configIntrospectNodes = append(configIntrospectNodes, nodesPortStr)
		}
		hostname := podList.Items[idx].Annotations["hostname"]
		statusMonitorConfig, err := StatusMonitorConfig(hostname, configIntrospectNodes,
			podList.Items[idx].Status.PodIP, "config", request.Name, request.Namespace, pod.Name)
		if err != nil {
			return err
		}
		data["monitorconfig."+podList.Items[idx].Status.PodIP+".yaml"] = statusMonitorConfig
		var configApiConfigBuffer bytes.Buffer
		configtemplates.ConfigAPIConfig.Execute(&configApiConfigBuffer, struct {
			HostIP              string
			ListenPort          string
			CassandraServerList string
			ZookeeperServerList string
			RabbitmqServerList  string
			CollectorServerList string
			RabbitmqUser        string
			RabbitmqPassword    string
			RabbitmqVhost       string
			AuthMode            AuthenticationMode
			AAAMode             AAAMode
			LogLevel            string
			CAFilePath          string
			ApiIntrospectPort   string
		}{
			HostIP:              podList.Items[idx].Status.PodIP,
			ListenPort:          strconv.Itoa(*configConfig.APIPort),
			CassandraServerList: cassandraNodesInformation.Endpoint,
			ZookeeperServerList: zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:  rabbitmqNodesInformation.ServerListCommaSeparatedSSL,
			CollectorServerList: collectorServerList,
			RabbitmqUser:        rabbitmqSecretUser,
			RabbitmqPassword:    rabbitmqSecretPassword,
			RabbitmqVhost:       rabbitmqSecretVhost,
			AuthMode:            configConfig.AuthMode,
			AAAMode:             configConfig.AAAMode,
			LogLevel:            configConfig.LogLevel,
			CAFilePath:          certificates.SignerCAFilepath,
			ApiIntrospectPort:   strconv.Itoa(*configConfig.ApiIntrospectPort),
		})
		data["api."+podList.Items[idx].Status.PodIP] = configApiConfigBuffer.String()

		var vncApiConfigBuffer bytes.Buffer
		configtemplates.ConfigAPIVNC.Execute(&vncApiConfigBuffer, struct {
			HostIP                 string
			ListenPort             string
			AuthMode               AuthenticationMode
			CAFilePath             string
			KeystoneAddress        string
			KeystonePort           int
			KeystoneUserDomainName string
			KeystoneAuthProtocol   string
		}{
			HostIP:                 podList.Items[idx].Status.PodIP,
			ListenPort:             strconv.Itoa(*configConfig.APIPort),
			AuthMode:               configConfig.AuthMode,
			CAFilePath:             certificates.SignerCAFilepath,
			KeystoneAddress:        configAuth.Address,
			KeystonePort:           configAuth.Port,
			KeystoneUserDomainName: configAuth.UserDomainName,
			KeystoneAuthProtocol:   configAuth.AuthProtocol,
		})
		data["vnc."+podList.Items[idx].Status.PodIP] = vncApiConfigBuffer.String()

		fabricMgmtIP := podList.Items[idx].Status.PodIP
		if c.Spec.ServiceConfiguration.FabricMgmtIP != "" {
			fabricMgmtIP = c.Spec.ServiceConfiguration.FabricMgmtIP
		}
		var configDevicemanagerConfigBuffer bytes.Buffer
		configtemplates.ConfigDeviceManagerConfig.Execute(&configDevicemanagerConfigBuffer, struct {
			HostIP                      string
			ApiServerList               string
			AnalyticsServerList         string
			CassandraServerList         string
			ZookeeperServerList         string
			RabbitmqServerList          string
			CollectorServerList         string
			RabbitmqUser                string
			RabbitmqPassword            string
			RabbitmqVhost               string
			LogLevel                    string
			FabricMgmtIP                string
			CAFilePath                  string
			DeviceManagerIntrospectPort string
			DMRunMode                   string
		}{
			HostIP:                      podList.Items[idx].Status.PodIP,
			ApiServerList:               apiServerList,
			AnalyticsServerList:         analyticsServerList,
			CassandraServerList:         cassandraNodesInformation.Endpoint,
			ZookeeperServerList:         zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:          rabbitmqNodesInformation.ServerListCommaSeparatedSSL,
			CollectorServerList:         collectorServerList,
			RabbitmqUser:                rabbitmqSecretUser,
			RabbitmqPassword:            rabbitmqSecretPassword,
			RabbitmqVhost:               rabbitmqSecretVhost,
			LogLevel:                    configConfig.LogLevel,
			FabricMgmtIP:                fabricMgmtIP,
			CAFilePath:                  certificates.SignerCAFilepath,
			DeviceManagerIntrospectPort: strconv.Itoa(*configConfig.DeviceManagerIntrospectPort),
			DMRunMode:                   DMRunModeFull,
		})
		data["devicemanager."+podList.Items[idx].Status.PodIP] = configDevicemanagerConfigBuffer.String()

		var fabricAnsibleConfigBuffer bytes.Buffer
		configtemplates.FabricAnsibleConf.Execute(&fabricAnsibleConfigBuffer, struct {
			HostIP              string
			CollectorServerList string
			LogLevel            string
			CAFilePath          string
		}{
			HostIP:              podList.Items[idx].Status.PodIP,
			CollectorServerList: collectorServerList,
			LogLevel:            configConfig.LogLevel,
			CAFilePath:          certificates.SignerCAFilepath,
		})
		data["contrail-fabric-ansible.conf."+podList.Items[idx].Status.PodIP] = fabricAnsibleConfigBuffer.String()

		var configKeystoneAuthConfBuffer bytes.Buffer
		configtemplates.ConfigKeystoneAuthConf.Execute(&configKeystoneAuthConfBuffer, struct {
			AdminUsername             string
			AdminPassword             string
			KeystoneAddress           string
			KeystonePort              int
			KeystoneAuthProtocol      string
			KeystoneUserDomainName    string
			KeystoneProjectDomainName string
			KeystoneRegion            string
			CAFilePath                string
		}{
			AdminUsername:             configAuth.AdminUsername,
			AdminPassword:             configAuth.AdminPassword,
			KeystoneAddress:           configAuth.Address,
			KeystonePort:              configAuth.Port,
			KeystoneAuthProtocol:      configAuth.AuthProtocol,
			KeystoneUserDomainName:    configAuth.UserDomainName,
			KeystoneProjectDomainName: configAuth.ProjectDomainName,
			KeystoneRegion:            configAuth.Region,
			CAFilePath:                certificates.SignerCAFilepath,
		})
		data["contrail-keystone-auth.conf"] = configKeystoneAuthConfBuffer.String()

		data["dnsmasq."+podList.Items[idx].Status.PodIP] = configtemplates.ConfigDNSMasqConfig

		var configSchematransformerConfigBuffer bytes.Buffer
		configtemplates.ConfigSchematransformerConfig.Execute(&configSchematransformerConfigBuffer, struct {
			HostIP               string
			ApiServerList        string
			AnalyticsServerList  string
			CassandraServerList  string
			ZookeeperServerList  string
			RabbitmqServerList   string
			CollectorServerList  string
			RabbitmqUser         string
			RabbitmqPassword     string
			RabbitmqVhost        string
			LogLevel             string
			CAFilePath           string
			SchemaIntrospectPort string
		}{
			HostIP:               podList.Items[idx].Status.PodIP,
			ApiServerList:        apiServerList,
			AnalyticsServerList:  analyticsServerList,
			CassandraServerList:  cassandraNodesInformation.Endpoint,
			ZookeeperServerList:  zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:   rabbitmqNodesInformation.ServerListCommaSeparatedSSL,
			CollectorServerList:  collectorServerList,
			RabbitmqUser:         rabbitmqSecretUser,
			RabbitmqPassword:     rabbitmqSecretPassword,
			RabbitmqVhost:        rabbitmqSecretVhost,
			LogLevel:             configConfig.LogLevel,
			CAFilePath:           certificates.SignerCAFilepath,
			SchemaIntrospectPort: strconv.Itoa(*configConfig.SchemaIntrospectPort),
		})
		data["schematransformer."+podList.Items[idx].Status.PodIP] = configSchematransformerConfigBuffer.String()

		var configServicemonitorConfigBuffer bytes.Buffer
		configtemplates.ConfigServicemonitorConfig.Execute(&configServicemonitorConfigBuffer, struct {
			HostIP                   string
			ApiServerList            string
			AnalyticsServerList      string
			CassandraServerList      string
			ZookeeperServerList      string
			RabbitmqServerList       string
			CollectorServerList      string
			RabbitmqUser             string
			RabbitmqPassword         string
			RabbitmqVhost            string
			AAAMode                  AAAMode
			LogLevel                 string
			CAFilePath               string
			SvcMonitorIntrospectPort string
		}{
			HostIP:                   podList.Items[idx].Status.PodIP,
			ApiServerList:            apiServerList,
			AnalyticsServerList:      analyticsServerSpaceSeparatedList,
			CassandraServerList:      cassandraNodesInformation.Endpoint,
			ZookeeperServerList:      zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:       rabbitmqNodesInformation.ServerListCommaSeparatedSSL,
			CollectorServerList:      collectorServerList,
			RabbitmqUser:             rabbitmqSecretUser,
			RabbitmqPassword:         rabbitmqSecretPassword,
			RabbitmqVhost:            rabbitmqSecretVhost,
			AAAMode:                  configConfig.AAAMode,
			LogLevel:                 configConfig.LogLevel,
			CAFilePath:               certificates.SignerCAFilepath,
			SvcMonitorIntrospectPort: strconv.Itoa(*configConfig.SvcMonitorIntrospectPort),
		})
		data["servicemonitor."+podList.Items[idx].Status.PodIP] = configServicemonitorConfigBuffer.String()

		var configAnalyticsapiConfigBuffer bytes.Buffer
		configtemplates.ConfigAnalyticsapiConfig.Execute(&configAnalyticsapiConfigBuffer, struct {
			HostIP                     string
			ApiServerList              string
			AnalyticsServerList        string
			CassandraServerList        string
			ZookeeperServerList        string
			RabbitmqServerList         string
			CollectorServerList        string
			RedisServerList            string
			RabbitmqUser               string
			RabbitmqPassword           string
			RabbitmqVhost              string
			AuthMode                   string
			AAAMode                    AAAMode
			CAFilePath                 string
			AnalyticsApiIntrospectPort string
		}{
			HostIP:                     podList.Items[idx].Status.PodIP,
			ApiServerList:              apiServerSpaceSeparatedList,
			AnalyticsServerList:        analyticsServerSpaceSeparatedList,
			CassandraServerList:        cassandraNodesInformation.Endpoint,
			ZookeeperServerList:        zookeeperNodesInformation.ServerListSpaceSeparated,
			RabbitmqServerList:         rabbitmqNodesInformation.ServerListCommaSeparatedSSL,
			CollectorServerList:        collectorServerList,
			RedisServerList:            redisServerSpaceSeparatedList,
			RabbitmqUser:               rabbitmqSecretUser,
			RabbitmqPassword:           rabbitmqSecretPassword,
			RabbitmqVhost:              rabbitmqSecretVhost,
			AAAMode:                    configConfig.AAAMode,
			CAFilePath:                 certificates.SignerCAFilepath,
			AnalyticsApiIntrospectPort: strconv.Itoa(*configConfig.AnalyticsApiIntrospectPort),
		})
		data["analyticsapi."+podList.Items[idx].Status.PodIP] = configAnalyticsapiConfigBuffer.String()
		/*
			command := []string{"/bin/sh", "-c", "hostname"}
			hostname, _, err := ExecToPodThroughAPI(command, "init", podList.Items[idx].Name, podList.Items[idx].Namespace, nil)
			if err != nil {
				return err
			}
		*/
		var configCollectorConfigBuffer bytes.Buffer
		configtemplates.ConfigCollectorConfig.Execute(&configCollectorConfigBuffer, struct {
			Hostname                string
			HostIP                  string
			ApiServerList           string
			CassandraServerList     string
			ZookeeperServerList     string
			RabbitmqServerList      string
			RabbitmqUser            string
			RabbitmqPassword        string
			RabbitmqVhost           string
			LogLevel                string
			CAFilePath              string
			CollectorIntrospectPort string
			AnalyticsDataTTL        string
			AnalyticsConfigAuditTTL string
			AnalyticsStatisticsTTL  string
			AnalyticsFlowTTL        string
		}{
			Hostname:                hostname,
			HostIP:                  podList.Items[idx].Status.PodIP,
			ApiServerList:           apiServerSpaceSeparatedList,
			CassandraServerList:     cassandraNodesInformation.ServerListCQLSpaceSeparated,
			ZookeeperServerList:     zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:      rabbitmqNodesInformation.ServerListSpaceSeparatedSSL,
			RabbitmqUser:            rabbitmqSecretUser,
			RabbitmqPassword:        rabbitmqSecretPassword,
			RabbitmqVhost:           rabbitmqSecretVhost,
			LogLevel:                configConfig.LogLevel,
			CAFilePath:              certificates.SignerCAFilepath,
			CollectorIntrospectPort: strconv.Itoa(*configConfig.CollectorIntrospectPort),
			AnalyticsDataTTL:        strconv.Itoa(*configConfig.AnalyticsDataTTL),
			AnalyticsConfigAuditTTL: strconv.Itoa(*configConfig.AnalyticsConfigAuditTTL),
			AnalyticsStatisticsTTL:  strconv.Itoa(*configConfig.AnalyticsStatisticsTTL),
			AnalyticsFlowTTL:        strconv.Itoa(*configConfig.AnalyticsFlowTTL),
		})
		data["collector."+podList.Items[idx].Status.PodIP] = configCollectorConfigBuffer.String()

		var configQueryEngineConfigBuffer bytes.Buffer
		configtemplates.ConfigQueryEngineConfig.Execute(&configQueryEngineConfigBuffer, struct {
			Hostname            string
			HostIP              string
			CassandraServerList string
			CollectorServerList string
			RedisServerList     string
			CAFilePath          string
			AnalyticsDataTTL    string
		}{
			Hostname:            hostname,
			HostIP:              podList.Items[idx].Status.PodIP,
			CassandraServerList: cassandraNodesInformation.ServerListCQLSpaceSeparated,
			CollectorServerList: collectorServerList,
			RedisServerList:     redisServerSpaceSeparatedList,
			CAFilePath:          certificates.SignerCAFilepath,
			AnalyticsDataTTL:    strconv.Itoa(*configConfig.AnalyticsDataTTL),
		})
		data["queryengine."+podList.Items[idx].Status.PodIP] = configQueryEngineConfigBuffer.String()

		var configNodemanagerconfigConfigBuffer bytes.Buffer
		configtemplates.ConfigNodemanagerConfigConfig.Execute(&configNodemanagerconfigConfigBuffer, struct {
			HostIP              string
			CollectorServerList string
			CassandraPort       string
			CassandraJmxPort    string
			CAFilePath          string
		}{
			HostIP:              podList.Items[idx].Status.PodIP,
			CollectorServerList: collectorServerList,
			CassandraPort:       cassandraNodesInformation.CQLPort,
			CassandraJmxPort:    cassandraNodesInformation.JMXPort,
			CAFilePath:          certificates.SignerCAFilepath,
		})
		data["nodemanagerconfig."+podList.Items[idx].Status.PodIP] = configNodemanagerconfigConfigBuffer.String()

		var configNodemanageranalyticsConfigBuffer bytes.Buffer
		configtemplates.ConfigNodemanagerAnalyticsConfig.Execute(&configNodemanageranalyticsConfigBuffer, struct {
			HostIP              string
			CollectorServerList string
			CassandraPort       string
			CassandraJmxPort    string
			CAFilePath          string
		}{
			HostIP:              podList.Items[idx].Status.PodIP,
			CollectorServerList: collectorServerList,
			CassandraPort:       cassandraNodesInformation.CQLPort,
			CassandraJmxPort:    cassandraNodesInformation.JMXPort,
			CAFilePath:          certificates.SignerCAFilepath,
		})
		data["nodemanageranalytics."+podList.Items[idx].Status.PodIP] = configNodemanageranalyticsConfigBuffer.String()
	}
	data["predef.json"] = predef
	configMapInstanceDynamicConfig.Data = data
	err = client.Update(context.TODO(), configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}

	return nil
}

type ConfigAuthParameters struct {
	AdminUsername     string
	AdminPassword     string
	Address           string
	Port              int
	Region            string
	AuthProtocol      string
	UserDomainName    string
	ProjectDomainName string
}

func (c *Config) AuthParameters(client client.Client) (*ConfigAuthParameters, error) {
	w := &ConfigAuthParameters{
		AdminUsername: "admin",
	}
	adminPasswordSecretName := c.Spec.ServiceConfiguration.KeystoneSecretName
	adminPasswordSecret := &corev1.Secret{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: adminPasswordSecretName, Namespace: c.Namespace}, adminPasswordSecret); err != nil {
		return nil, err
	}
	w.AdminPassword = string(adminPasswordSecret.Data["password"])

	if c.Spec.ServiceConfiguration.AuthMode == AuthenticationModeKeystone {
		keystoneInstanceName := c.Spec.ServiceConfiguration.KeystoneInstance
		keystone := &Keystone{}
		if err := client.Get(context.TODO(), types.NamespacedName{Namespace: c.Namespace, Name: keystoneInstanceName}, keystone); err != nil {
			return nil, err
		}
		if keystone.Status.Endpoint == "" {
			return nil, fmt.Errorf("%q Status.Endpoint empty", keystoneInstanceName)
		}
		w.Port = keystone.Spec.ServiceConfiguration.ListenPort
		w.Region = keystone.Spec.ServiceConfiguration.Region
		w.AuthProtocol = keystone.Spec.ServiceConfiguration.AuthProtocol
		w.UserDomainName = keystone.Spec.ServiceConfiguration.UserDomainName
		w.ProjectDomainName = keystone.Spec.ServiceConfiguration.ProjectDomainName
		w.Address = keystone.Status.Endpoint
	}

	return w, nil
}

func (c *Config) CreateConfigMap(configMapName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (*corev1.ConfigMap, error) {
	return CreateConfigMap(configMapName,
		client,
		scheme,
		request,
		"config",
		c)
}

// CurrentConfigMapExists checks if a current configuration exists and returns it.
func (c *Config) CurrentConfigMapExists(configMapName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (corev1.ConfigMap, bool) {
	return CurrentConfigMapExists(configMapName,
		client,
		scheme,
		request)
}

// CreateSecret creates a secret.
func (c *Config) CreateSecret(secretName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (*corev1.Secret, error) {
	return CreateSecret(secretName,
		client,
		scheme,
		request,
		"config",
		c)
}

// PrepareSTS prepares the intented statefulset for the config object
func (c *Config) PrepareSTS(sts *appsv1.StatefulSet, commonConfiguration *PodConfiguration, request reconcile.Request, scheme *runtime.Scheme, client client.Client) error {
	return PrepareSTS(sts, commonConfiguration, "config", request, scheme, c, client, true)
}

// AddVolumesToIntendedSTS adds volumes to the config statefulset
func (c *Config) AddVolumesToIntendedSTS(sts *appsv1.StatefulSet, volumeConfigMapMap map[string]string) {
	AddVolumesToIntendedSTS(sts, volumeConfigMapMap)
}

// AddSecretVolumesToIntendedSTS adds volumes to the Rabbitmq deployment.
func (c *Config) AddSecretVolumesToIntendedSTS(sts *appsv1.StatefulSet, volumeConfigMapMap map[string]string) {
	AddSecretVolumesToIntendedSTS(sts, volumeConfigMapMap)
}

//CreateSTS creates the STS
func (c *Config) CreateSTS(sts *appsv1.StatefulSet, instanceType string, request reconcile.Request, reconcileClient client.Client) error {
	return CreateSTS(sts, instanceType, request, reconcileClient)
}

//UpdateSTS updates the STS
func (c *Config) UpdateSTS(sts *appsv1.StatefulSet, instanceType string, request reconcile.Request, reconcileClient client.Client, strategy string) error {
	return UpdateSTS(sts, instanceType, request, reconcileClient, strategy)
}

// SetInstanceActive sets the Cassandra instance to active
func (c *Config) SetInstanceActive(client client.Client, activeStatus *bool, sts *appsv1.StatefulSet, request reconcile.Request) error {
	if err := client.Get(context.TODO(), types.NamespacedName{Name: sts.Name, Namespace: request.Namespace},
		sts); err != nil {
		return err
	}

	*activeStatus = false
	acceptableReadyReplicaCnt := int32(1)
	if sts.Spec.Replicas != nil {
		acceptableReadyReplicaCnt = *sts.Spec.Replicas/2 + 1
	}

	if sts.Status.ReadyReplicas >= acceptableReadyReplicaCnt {
		*activeStatus = true
	}

	if err := client.Status().Update(context.TODO(), c); err != nil {
		return err
	}
	return nil
}

// PodIPListAndIPMapFromInstance gets a list with POD IPs and a map of POD names and IPs.
func (c *Config) PodIPListAndIPMapFromInstance(request reconcile.Request, reconcileClient client.Client) (*corev1.PodList, map[string]string, error) {
	return PodIPListAndIPMapFromInstance("config", &c.Spec.CommonConfiguration, request, reconcileClient, true, true, false, false, false, false)
}

func (c *Config) SetPodsToReady(podIPList *corev1.PodList, client client.Client) error {
	return SetPodsToReady(podIPList, client)
}

func (c *Config) WaitForPeerPods(request reconcile.Request, reconcileClient client.Client) error {
	labelSelector := labels.SelectorFromSet(map[string]string{"config": request.Name})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	list := &corev1.PodList{}
	err := reconcileClient.List(context.TODO(), list, listOps)
	if err != nil {
		return err
	}
	sort.SliceStable(list.Items, func(i, j int) bool { return list.Items[i].Name < list.Items[j].Name })
	for idx, pod := range list.Items {
		ready := true
		for i := 0; i < idx; i++ {
			for _, containerStatus := range list.Items[i].Status.ContainerStatuses {
				if !containerStatus.Ready {
					ready = false
				}
			}
		}
		if ready {
			podTOUpdate := &corev1.Pod{}
			err := reconcileClient.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, podTOUpdate)
			if err != nil {
				return err
			}
			podTOUpdate.ObjectMeta.Labels["peers_ready"] = "true"
			err = reconcileClient.Update(context.TODO(), podTOUpdate)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) ManageNodeStatus(podNameIPMap map[string]string, client client.Client) error {
	c.Status.Nodes = podNameIPMap
	configConfigInterface := c.ConfigurationParameters()
	configConfig := configConfigInterface.(ConfigConfiguration)
	c.Status.Ports.APIPort = strconv.Itoa(*configConfig.APIPort)
	c.Status.Ports.AnalyticsPort = strconv.Itoa(*configConfig.AnalyticsPort)
	c.Status.Ports.CollectorPort = strconv.Itoa(*configConfig.CollectorPort)
	c.Status.Ports.RedisPort = strconv.Itoa(*configConfig.RedisPort)
	err := client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

// IsActive returns true if instance is active
func (c *Config) IsActive(name string, namespace string, myclient client.Client) bool {
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": name})
	listOps := &client.ListOptions{Namespace: namespace, LabelSelector: labelSelector}
	list := &ConfigList{}
	err := myclient.List(context.TODO(), list, listOps)
	if err != nil {
		return false
	}
	if len(list.Items) > 0 {
		if list.Items[0].Status.Active != nil {
			if *list.Items[0].Status.Active {
				return true
			}
		}
	}
	return false
}

func (c *Config) ConfigurationParameters() interface{} {
	configConfiguration := ConfigConfiguration{}
	var apiPort int
	var analyticsPort int
	var collectorPort int
	var redisPort int
	var rabbitmqUser string
	var rabbitmqPassword string
	var rabbitmqVhost string
	var logLevel string
	if c.Spec.ServiceConfiguration.LogLevel != "" {
		logLevel = c.Spec.ServiceConfiguration.LogLevel
	} else {
		logLevel = LogLevel
	}
	configConfiguration.LogLevel = logLevel
	if c.Spec.ServiceConfiguration.APIPort != nil {
		apiPort = *c.Spec.ServiceConfiguration.APIPort
	} else {
		apiPort = ConfigApiPort
	}
	configConfiguration.APIPort = &apiPort

	if c.Spec.ServiceConfiguration.AnalyticsPort != nil {
		analyticsPort = *c.Spec.ServiceConfiguration.AnalyticsPort
	} else {
		analyticsPort = AnalyticsApiPort
	}
	configConfiguration.AnalyticsPort = &analyticsPort

	if c.Spec.ServiceConfiguration.CollectorPort != nil {
		collectorPort = *c.Spec.ServiceConfiguration.CollectorPort
	} else {
		collectorPort = CollectorPort
	}
	configConfiguration.CollectorPort = &collectorPort

	if c.Spec.ServiceConfiguration.RedisPort != nil {
		redisPort = *c.Spec.ServiceConfiguration.RedisPort
	} else {
		redisPort = RedisServerPort
	}
	configConfiguration.RedisPort = &redisPort

	var apiIntrospectPort int
	if c.Spec.ServiceConfiguration.ApiIntrospectPort != nil {
		apiIntrospectPort = *c.Spec.ServiceConfiguration.ApiIntrospectPort
	} else {
		apiIntrospectPort = ConfigApiIntrospectPort
	}
	configConfiguration.ApiIntrospectPort = &apiIntrospectPort

	var schemaIntrospectPort int
	if c.Spec.ServiceConfiguration.SchemaIntrospectPort != nil {
		schemaIntrospectPort = *c.Spec.ServiceConfiguration.SchemaIntrospectPort
	} else {
		schemaIntrospectPort = ConfigSchemaIntrospectPort
	}
	configConfiguration.SchemaIntrospectPort = &schemaIntrospectPort

	var deviceManagerIntrospectPort int
	if c.Spec.ServiceConfiguration.DeviceManagerIntrospectPort != nil {
		deviceManagerIntrospectPort = *c.Spec.ServiceConfiguration.DeviceManagerIntrospectPort
	} else {
		deviceManagerIntrospectPort = ConfigDeviceManagerIntrospectPort
	}
	configConfiguration.DeviceManagerIntrospectPort = &deviceManagerIntrospectPort

	var svcMonitorIntrospectPort int
	if c.Spec.ServiceConfiguration.SvcMonitorIntrospectPort != nil {
		svcMonitorIntrospectPort = *c.Spec.ServiceConfiguration.SvcMonitorIntrospectPort
	} else {
		svcMonitorIntrospectPort = ConfigSvcMonitorIntrospectPort
	}
	configConfiguration.SvcMonitorIntrospectPort = &svcMonitorIntrospectPort

	var analyticsApiIntrospectPort int
	if c.Spec.ServiceConfiguration.AnalyticsApiIntrospectPort != nil {
		analyticsApiIntrospectPort = *c.Spec.ServiceConfiguration.AnalyticsApiIntrospectPort
	} else {
		analyticsApiIntrospectPort = AnalyticsApiIntrospectPort
	}
	configConfiguration.AnalyticsApiIntrospectPort = &analyticsApiIntrospectPort

	var collectorIntrospectPort int
	if c.Spec.ServiceConfiguration.CollectorIntrospectPort != nil {
		collectorIntrospectPort = *c.Spec.ServiceConfiguration.CollectorIntrospectPort
	} else {
		collectorIntrospectPort = CollectorIntrospectPort
	}
	configConfiguration.CollectorIntrospectPort = &collectorIntrospectPort

	if c.Spec.ServiceConfiguration.NodeManager != nil {
		configConfiguration.NodeManager = c.Spec.ServiceConfiguration.NodeManager
	} else {
		nodeManager := true
		configConfiguration.NodeManager = &nodeManager
	}

	if c.Spec.ServiceConfiguration.RabbitmqUser != "" {
		rabbitmqUser = c.Spec.ServiceConfiguration.RabbitmqUser
	} else {
		rabbitmqUser = RabbitmqUser
	}
	configConfiguration.RabbitmqUser = rabbitmqUser

	if c.Spec.ServiceConfiguration.RabbitmqPassword != "" {
		rabbitmqPassword = c.Spec.ServiceConfiguration.RabbitmqPassword
	} else {
		rabbitmqPassword = RabbitmqPassword
	}
	configConfiguration.RabbitmqPassword = rabbitmqPassword

	if c.Spec.ServiceConfiguration.RabbitmqVhost != "" {
		rabbitmqVhost = c.Spec.ServiceConfiguration.RabbitmqVhost
	} else {
		rabbitmqVhost = RabbitmqVhost
	}
	configConfiguration.RabbitmqVhost = rabbitmqVhost

	configConfiguration.AuthMode = c.Spec.ServiceConfiguration.AuthMode
	if configConfiguration.AuthMode == "" {
		configConfiguration.AuthMode = AuthenticationModeNoAuth
	}

	configConfiguration.AAAMode = c.Spec.ServiceConfiguration.AAAMode
	if configConfiguration.AAAMode == "" {
		configConfiguration.AAAMode = AAAModeNoAuth
		if configConfiguration.AuthMode == AuthenticationModeKeystone {
			configConfiguration.AAAMode = AAAModeRBAC
		}
	}

	var analyticsDataTTL int
	if c.Spec.ServiceConfiguration.AnalyticsDataTTL != nil {
		analyticsDataTTL = *c.Spec.ServiceConfiguration.AnalyticsDataTTL
	} else {
		analyticsDataTTL = AnalyticsDataTTL
	}
	configConfiguration.AnalyticsDataTTL = &analyticsDataTTL

	var analyticsConfigAuditTTL int
	if c.Spec.ServiceConfiguration.AnalyticsConfigAuditTTL != nil {
		analyticsConfigAuditTTL = *c.Spec.ServiceConfiguration.AnalyticsConfigAuditTTL
	} else {
		analyticsConfigAuditTTL = AnalyticsConfigAuditTTL
	}
	configConfiguration.AnalyticsConfigAuditTTL = &analyticsConfigAuditTTL

	var analyticsStatisticsTTL int
	if c.Spec.ServiceConfiguration.AnalyticsStatisticsTTL != nil {
		analyticsStatisticsTTL = *c.Spec.ServiceConfiguration.AnalyticsStatisticsTTL
	} else {
		analyticsStatisticsTTL = AnalyticsStatisticsTTL
	}
	configConfiguration.AnalyticsStatisticsTTL = &analyticsStatisticsTTL

	var analyticsFlowTTL int
	if c.Spec.ServiceConfiguration.AnalyticsFlowTTL != nil {
		analyticsFlowTTL = *c.Spec.ServiceConfiguration.AnalyticsFlowTTL
	} else {
		analyticsFlowTTL = AnalyticsFlowTTL
	}
	configConfiguration.AnalyticsFlowTTL = &analyticsFlowTTL

	return configConfiguration

}

func (c *Config) SetEndpointInStatus(client client.Client, clusterIP string) error {
	c.Status.Endpoint = clusterIP
	err := client.Status().Update(context.TODO(), c)
	return err
}

var predef = `{
  "data": [
    {
      "object_type": "job-template",
      "objects": [
        {
          "fq_name": [
            "default-global-system-config",
            "image_upgrade_template"
          ],
          "name": "image_upgrade_template",
          "display_name": "Image Upgrade",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "device",
          "job_template_type": "workflow",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": true,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/image_upgrade.yml",
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_input_schema": "",
          "job_template_output_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "container_cleanup_template"
          ],
          "name": "container_cleanup_template",
          "display_name": "Container Cleanup",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "fabric",
          "job_template_type": "workflow",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/container_cleanup.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "device_family": ""
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "cli_sync_template"
          ],
          "name": "cli_sync_template",
          "display_name": "Sync cli config with Contrail",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "fabric",
          "job_template_type": "workflow",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/cli_sync.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": ""
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "hitless_upgrade_strategy_template"
          ],
          "name": "hitless_upgrade_strategy_template",
          "display_name": "Hitless Image Upgrade",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "fabric",
          "job_template_type": "workflow",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/hitless_upgrade_strategy.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 10,
                "sequence_no": 0
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/hitless_upgrade.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 90,
                "sequence_no": 1
              }

            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "maintenance_mode_activate_template"
          ],
          "name": "maintenance_mode_activate_template",
          "display_name": "Maintenance Mode Activation",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "device",
          "job_template_type": "workflow",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": false,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/maintenance_mode_activate.yml",
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_input_schema": "",
          "job_template_output_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "maintenance_mode_deactivate_template"
          ],
          "name": "maintenance_mode_deactivate_template",
          "display_name": "Maintenance Mode Deactivation",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "device",
          "job_template_type": "workflow",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": false,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/maintenance_mode_deactivate.yml",
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_input_schema": "",
          "job_template_output_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "discover_device_template"
          ],
          "name": "discover_device_template",
          "display_name": "Discover Device",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "fabric",
          "job_template_type": "workflow",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": false,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/discover_device.yml",
                "vendor": ""
              }
            ]
          },
          "job_template_input_schema": "",
          "job_template_output_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "discover_os_computes_template"
          ],
          "name": "discover_os_computes_template",
          "display_name": "Discover OS Computes",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": false,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/discover_os_computes_template.yml",
                "vendor": ""
              }
            ]
          },
          "job_template_input_schema": "",
          "job_template_output_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "discover_role_template"
          ],
          "name": "discover_role_template",
          "display_name": "Discover Role",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "fabric",
          "job_template_type": "workflow",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": true,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/discover_role.yml",
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "device_import_template"
          ],
          "name": "device_import_template",
          "display_name": "Device Import",
          "parent_type": "global-system-config",
          "job_template_type": "workflow",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/device_import.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 50,
                "sequence_no": 0
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/topology_discovery.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 50,
                "sequence_no": 1
              }            
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "fabric_config_template"
          ],
          "name": "fabric_config_template",
          "display_name": "Device Config Push",
          "parent_type": "global-system-config",
          "job_template_type": "config",
          "job_template_concurrency_level": "device",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/fabric_config.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": ""
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "ztp_template"
          ],
          "name": "ztp_template",
          "display_name": "ZTP",
          "parent_type": "global-system-config",
          "job_template_type": "workflow",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/ztp.yml",
                "multi_device_playbook": false,
                "vendor": ""
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "fabric_deletion_template"
          ],
          "name": "fabric_deletion_template",
          "display_name": "Fabric Deletion",
          "parent_type": "global-system-config",
          "job_template_type": "workflow",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/delete_fabric.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "device_family": ""
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "device_deletion_template"
          ],
          "name": "device_deletion_template",
          "display_name": "Device Deletion",
          "parent_type": "global-system-config",
          "job_template_type": "workflow",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/delete_fabric_devices.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 50,
                "sequence_no": 0
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/update_dhcp_config.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 50,
                "sequence_no": 1
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "role_assignment_template"
          ],
          "name": "role_assignment_template",
          "display_name": "Role Assignment",
          "parent_type": "global-system-config",
          "job_template_type": "config",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/role_assignment.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "device_family": ""
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "topology_discovery_template"
          ],
          "name": "topology_discovery_template",
          "display_name": "Topology Discovery",
          "parent_type": "global-system-config",
          "job_template_type": "workflow",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/topology_discovery.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "fabric_onboard_template"
          ],
          "name": "fabric_onboard_template",
          "display_name": "Fabric Onboarding",
          "parent_type": "global-system-config",
          "job_template_type": "workflow",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/fabric_onboard.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 5,
                "sequence_no": 0
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/ztp.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 15,
                "sequence_no": 1
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/discover_device.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 15,
                "sequence_no": 2
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/assign_static_device_ip.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 10,
                "sequence_no": 3
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/ztp_select_image.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 15,
                "sequence_no": 4
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/role_assignment_dfg.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 5,
                "sequence_no": 5
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/discover_role.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 5,
                "sequence_no": 6
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/hardware_inventory.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 5,
                "sequence_no": 7
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/device_import.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 10,
                "sequence_no": 8
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/topology_discovery.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 10,
                "sequence_no": 9
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/update_dhcp_config.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 5,
                "sequence_no": 10
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "search_ip_mac_template"
          ],
          "name": "search_ip_mac_template",
          "display_name": "Search using IP or MAC",
          "job_template_description": "This command is used to locate an interface in a fabric given an IP or MAC",
          "parent_type": "global-system-config",
          "job_template_type": "device_operation",
          "job_template_concurrency_level": "device",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": true,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/operational_command.yml",
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": "",
          "job_template_input_ui_schema": "",
          "job_template_output_ui_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "show_interface_details_template"
          ],
          "name": "show_interface_details_template",
          "display_name": "Show Interface Details",
          "job_template_description": "This command is used to get a list of physical or logical interfaces from a device and interfaces' other related information.",
          "parent_type": "global-system-config",
          "job_template_type": "device_operation",
          "job_template_concurrency_level": "device",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": true,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/operational_command.yml",
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": "",
          "job_template_input_ui_schema": "",
          "job_template_output_ui_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "show_config_template"
          ],
          "name": "show_config_template",
          "display_name": " Show current or rollback configuration",
          "job_template_description": "This command is used to display current or previous configuration",
          "parent_type": "global-system-config",
          "job_template_type": "device_operation",
          "job_template_concurrency_level": "device",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": true,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/operational_command.yml",
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": "",
          "job_template_input_ui_schema": "",
          "job_template_output_ui_schema": ""
        },
        {
            "fq_name": [
              "default-global-system-config",
              "test_overlay_connectivity_template"
            ],
            "name": "test_overlay_connectivity_template",
            "display_name": "Test overlay connectivity",
            "job_template_description": "",
            "parent_type": "global-system-config",
            "job_template_type": "device_operation",
            "job_template_concurrency_level": "device",
            "job_template_playbooks": {
              "playbook_info": [
                {
                  "device_family": "",
                  "multi_device_playbook": true,
                  "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/operational_command.yml",
                  "vendor": "Juniper"
                }
              ]
            },
            "job_template_output_schema": "",
            "job_template_input_schema": "",
            "job_template_input_ui_schema": "",
            "job_template_output_ui_schema": ""
        },
        {
         "fq_name": [
            "default-global-system-config",
            "show_ops_info_template"
          ],
          "name": "show_ops_info_template",
          "display_name": "Show operations information",
          "job_template_description": "This command is used to display operational information",
          "parent_type": "global-system-config",
          "job_template_type": "device_operation",
          "job_template_concurrency_level": "device",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": true,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/operational_command.yml",
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": "",
          "job_template_input_ui_schema": "",
          "job_template_output_ui_schema": ""
        },
        {
            "fq_name": [
               "default-global-system-config",
               "check_multicast_template"
             ],
             "name": "check_multicast_template",
             "display_name": "Check incoming multicast traffic",
             "job_template_description": "Check multicast",
             "parent_type": "global-system-config",
             "job_template_type": "device_operation",
             "job_template_concurrency_level": "device",
             "job_template_playbooks": {
               "playbook_info": [
                 {
                   "device_family": "",
                   "multi_device_playbook": true,
                   "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/operational_command.yml",
                   "vendor": "Juniper"
                 }
               ]
             },
             "job_template_output_schema": "",
             "job_template_input_schema": "",
             "job_template_input_ui_schema": "",
             "job_template_output_ui_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "existing_fabric_onboard_template"
          ],
          "name": "existing_fabric_onboard_template",
          "display_name": "Existing Fabric Onboarding",
          "parent_type": "global-system-config",
          "job_template_type": "workflow",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/existing_fabric_onboard.yml",
                "vendor": "Juniper",
                "multi_device_playbook": false,
                "device_family": "",
                "job_completion_weightage": 10,
                "sequence_no": 0
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/discover_device.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 20,
                "sequence_no": 1
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/discover_role.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 10,
                "sequence_no": 2
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/hardware_inventory.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 5,
                "sequence_no": 3
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/device_import.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 40,
                "sequence_no": 4
              },
              {
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/topology_discovery.yml",
                "multi_device_playbook": true,
                "vendor": "Juniper",
                "device_family": "",
                "job_completion_weightage": 15,
                "sequence_no": 5
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "vcenter_import_template"
          ],
          "name": "vcenter_import_template",
          "display_name": "VCenter Import",
          "parent_type": "global-system-config",
          "job_template_type": "executable",
          "job_template_concurrency_level": "fabric",
          "job_template_executables": {
            "executable_info": [
              {
                "executable_path": "/opt/contrail/utils/vcenter-import",
                "executable_args": "",
                "job_completion_weightage": 100
              }
            ]
          },
          "job_template_input_schema": "",
          "job_template_output_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "discover_server_template"
          ],
          "name": "discover_server_template",
          "display_name": "Server Discovery",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/discover_server.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "job_completion_weightage": 100
              }
            ]
          },
          "job_template_input_schema": "",
          "job_template_output_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "server_import_template"
          ],
          "name": "server_import_template",
          "display_name": "Server Import",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/server_import.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "job_completion_weightage": 100
              }
            ]
          },
          "job_template_input_schema": "",
          "job_template_output_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "node_profile_template"
          ],
          "name": "node_profile_template",
          "display_name": "Node Profile",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "fabric",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/node_profile.yml",
                "multi_device_playbook": false,
                "vendor": "Juniper",
                "job_completion_weightage": 100
              }
            ]
          },
          "job_template_input_schema": "",
          "job_template_output_schema": ""
        },
        {
         "fq_name": [
            "default-global-system-config",
            "show_mac_mob_template"
          ],
          "name": "show_mac_mob_template",
          "display_name": "Show mac mobility",
          "job_template_description": "This command is used to display MAC address move or mobility",
          "parent_type": "global-system-config",
          "job_template_type": "device_operation",
          "job_template_concurrency_level": "device",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": true,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/operational_command.yml",
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": "",
          "job_template_input_ui_schema": "",
          "job_template_output_ui_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "rma_activate_template"
          ],
          "name": "rma_activate_template",
          "display_name": "RMA Activation",
          "parent_type": "global-system-config",
          "job_template_concurrency_level": "fabric",
          "job_template_type": "workflow",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": false,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/rma_activate.yml",
                "vendor": "Juniper",
                "job_completion_weightage": 15,
                "sequence_no": 0
              },
              {
                "device_family": "",
                "multi_device_playbook": true,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/assign_static_device_ip.yml",
                "vendor": "Juniper",
                "job_completion_weightage": 15,
                "sequence_no": 1
              },
              {
                "device_family": "",
                "multi_device_playbook": true,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/ztp_select_image.yml",
                "vendor": "Juniper",
                "job_completion_weightage": 50,
                "sequence_no": 2
              },
              {
                "device_family": "",
                "multi_device_playbook": false,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/update_dhcp_config.yml",
                "vendor": "Juniper",
                "job_completion_weightage": 20,
                "sequence_no": 3
              }
            ]
          },
          "job_template_input_schema": "",
          "job_template_output_schema": ""
        },
        {
         "fq_name": [
            "default-global-system-config",
            "show_chassis_info_template"
          ],
          "name": "show_chassis_info_template",
          "display_name": "Show chassis information",
          "job_template_description": "This command is used to display chassis information",
          "parent_type": "global-system-config",
          "job_template_type": "device_operation",
          "job_template_concurrency_level": "device",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": true,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/operational_command.yml",
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": "",
          "job_template_input_ui_schema": "",
          "job_template_output_ui_schema": ""
        },
        {
          "fq_name": [
            "default-global-system-config",
            "hardware_inventory_template"
          ],
          "name": "hardware_inventory_template",
          "display_name": "Fetch hardware chassis information",
          "job_template_description": "This command is used to fetch hardware inventory information from the device",
          "parent_type": "global-system-config",
          "job_template_type": "workflow",
          "job_template_concurrency_level": "device",
          "job_template_playbooks": {
            "playbook_info": [
              {
                "device_family": "",
                "multi_device_playbook": true,
                "playbook_uri": "./opt/contrail/fabric_ansible_playbooks/hardware_inventory.yml",
                "vendor": "Juniper"
              }
            ]
          },
          "job_template_output_schema": "",
          "job_template_input_schema": "",
          "job_template_input_ui_schema": "",
          "job_template_output_ui_schema": ""
        }
      ]
    },
    {
      "object_type": "tag",
      "objects": [
        {
          "fq_name": [
            "label=fabric-management-ip"
          ],
          "name": "label=fabric-management-ip",
          "tag_type_name": "label",
          "tag_value": "fabric-management-ip",
          "tag_predefined": true
        },
        {
          "fq_name": [
            "label=fabric-loopback-ip"
          ],
          "name": "label=fabric-loopback-ip",
          "tag_type_name": "label",
          "tag_value": "fabric-loopback-ip",
          "tag_predefined": true
        },
        {
          "fq_name": [
            "label=fabric-overlay-loopback-ip"
          ],
          "name": "label=fabric-overlay-loopback-ip",
          "tag_type_name": "label",
          "tag_value": "fabric-overlay-loopback-ip",
          "tag_predefined": true
        },
        {
          "fq_name": [
            "label=fabric-peer-ip"
          ],
          "name": "label=fabric-peer-ip",
          "tag_type_name": "label",
          "tag_value": "fabric-peer-ip",
          "tag_predefined": true
        },
        {
          "fq_name": [
            "label=fabric-pnf-servicechain-ip"
          ],
          "name": "label=fabric-pnf-servicechain-ip",
          "tag_type_name": "label",
          "tag_value": "fabric-pnf-servicechain-ip",
          "tag_predefined": true
        },
        {
          "fq_name": [
            "label=fabric-as-number"
          ],
          "name": "label=fabric-as-number",
          "tag_type_name": "label",
          "tag_value": "fabric-as-number",
          "tag_predefined": true
        },
        {
          "fq_name": [
            "label=fabric-ebgp-as-number"
          ],
          "name": "label=fabric-ebgp-as-number",
          "tag_type_name": "label",
          "tag_value": "fabric-ebgp-as-number",
          "tag_predefined": true
        }
      ]
    },
    {
      "object_type": "global-system-config",
      "objects": [
        {
          "fq_name": [
            "default-global-system-config"
          ],
          "supported_device_families": {
            "device_family": [
              "junos",
              "junos-qfx"
            ]
          },
          "supported_vendor_hardwares": {
            "vendor_hardware": [
              "juniper-qfx10002-72q",
              "juniper-qfx10002-36q",
              "juniper-qfx10002-60c",
              "juniper-qfx10016",
              "juniper-qfx10008",
              "juniper-vqfx-10000",
              "juniper-qfx5100-48sh-afi",
              "juniper-qfx5100-48sh-afo",
              "juniper-qfx5100-48th-afi",
              "juniper-qfx5100-48th-afo",
              "juniper-qfx5100-48s-6q",
              "juniper-qfx5100-96s-8q",
              "juniper-qfx5100-48t",
              "juniper-qfx5100-48t-6q",
              "juniper-qfx5100-24q-2p",
              "juniper-qfx5100-24q-aa",
              "juniper-qfx5100e-48s-6q",
              "juniper-qfx5100e-96s-8q",
              "juniper-qfx5100e-48t-6q",
              "juniper-qfx5100e-24q-2p",
              "juniper-qfx5110-48s-4c",
              "juniper-qfx5110-32q",
              "juniper-qfx5120-48y-8c",
              "juniper-qfx5120-48t-6c",
              "juniper-qfx5120-32c",
              "juniper-qfx5220-32cd",
              "juniper-qfx5220-128c",
              "juniper-qfx5200-32c-32q",
              "juniper-qfx5200-48y",
              "juniper-qfx5210-64c",
              "juniper-vmx",
              "juniper-mx80",
              "juniper-mx80-48t",
              "juniper-mx204",
              "juniper-mx240",
              "juniper-mx480",
              "juniper-mx960",
              "juniper-mx2008",
              "juniper-mx2010",
              "juniper-mx2020",
              "juniper-mx10003",
              "juniper-jnp10008",
              "juniper-jnp10016",
              "juniper-srx5400",
              "juniper-srx5600",
              "juniper-srx4600",
              "juniper-srx4100",
              "juniper-srx1500",
              "juniper-srx240h-poe",
              "juniper-vsrx",
              "juniper-srx5800",
              "juniper-srx4200"
            ]
          }
        }
      ]
    },
    {
      "object_type": "hardware",
      "objects": [
        {
          "fq_name": [
            "juniper-qfx5100-48sh-afi"
          ],
          "name": "juniper-qfx5100-48sh-afi"
        },
        {
          "fq_name": [
            "juniper-qfx5100-48sh-afo"
          ],
          "name": "juniper-qfx5100-48sh-afo"
        },
        {
          "fq_name": [
            "juniper-qfx5100-48th-afi"
          ],
          "name": "juniper-qfx5100-48th-afi"
        },
        {
          "fq_name": [
            "juniper-qfx5100-48th-afo"
          ],
          "name": "juniper-qfx5100-48th-afo"
        },
        {
          "fq_name": [
            "juniper-qfx5100-48s-6q"
          ],
          "name": "juniper-qfx5100-48s-6q"
        },
        {
          "fq_name": [
            "juniper-qfx5100-96s-8q"
          ],
          "name": "juniper-qfx5100-96s-8q"
        },
        {
          "fq_name": [
            "juniper-qfx5100-48t"
          ],
          "name": "juniper-qfx5100-48t"
        },
        {
          "fq_name": [
            "juniper-qfx5100-48t-6q"
          ],
          "name": "juniper-qfx5100-48t-6q"
        },
        {
          "fq_name": [
            "juniper-qfx5100-24q-2p"
          ],
          "name": "juniper-qfx5100-24q-2p"
        },
        {
          "fq_name": [
            "juniper-qfx5100-24q-aa"
          ],
          "name": "juniper-qfx5100-24q-aa"
        },
        {
          "fq_name": [
            "juniper-qfx5100e-48s-6q"
          ],
          "name": "juniper-qfx5100e-48s-6q"
        },
        {
          "fq_name": [
            "juniper-qfx5100e-96s-8q"
          ],
          "name": "juniper-qfx5100e-96s-8q"
        },
        {
          "fq_name": [
            "juniper-qfx5100e-48t-6q"
          ],
          "name": "juniper-qfx5100e-48t-6q"
        },
        {
          "fq_name": [
            "juniper-qfx5100e-24q-2p"
          ],
          "name": "juniper-qfx5100e-24q-2p"
        },
        {
          "fq_name": [
            "juniper-qfx5110-48s-4c"
          ],
          "name": "juniper-qfx5110-48s-4c"
        },
        {
          "fq_name": [
            "juniper-qfx5110-32q"
          ],
          "name": "juniper-qfx5110-32q"
        },
        {
          "fq_name": [
            "juniper-qfx5120-48y-8c"
          ],
          "name": "juniper-qfx5120-48y-8c"
        },
        {
          "fq_name": [
            "juniper-qfx5120-32c"
          ],
          "name": "juniper-qfx5120-32c"
        },
        {
          "fq_name": [
            "juniper-qfx5120-48t-6c"
          ],
          "name": "juniper-qfx5120-48t-6c"
        },
        {
          "fq_name": [
            "juniper-qfx5220-32cd"
          ],
          "name": "juniper-qfx5220-32cd"
        },
        {
          "fq_name": [
            "juniper-qfx5220-128c"
          ],
          "name": "juniper-qfx5220-128c"
        },
        {
          "fq_name": [
            "juniper-qfx5200-32c-32q"
          ],
          "name": "juniper-qfx5200-32c-32q"
        },
        {
          "fq_name": [
            "juniper-qfx5200-48y"
          ],
          "name": "juniper-qfx5200-48y"
        },
        {
          "fq_name": [
            "juniper-qfx5210-64c"
          ],
          "name": "juniper-qfx5210-64c"
        },
        {
          "fq_name": [
            "juniper-qfx10002-72q"
          ],
          "name": "juniper-qfx10002-72q"
        },
        {
          "fq_name": [
            "juniper-qfx10002-36q"
          ],
          "name": "juniper-qfx10002-36q"
        },
        {
          "fq_name": [
            "juniper-qfx10002-60c"
          ],
          "name": "juniper-qfx10002-60c"
        },
        {
          "fq_name": [
            "juniper-qfx10016"
          ],
          "name": "juniper-qfx10016"
        },
        {
          "fq_name": [
            "juniper-qfx10008"
          ],
          "name": "juniper-qfx10008"
        },
        {
          "fq_name": [
            "juniper-vqfx-10000"
          ],
          "name": "juniper-vqfx-10000"
        },
        {
          "fq_name": [
            "juniper-mx80"
          ],
          "name": "juniper-mx80"
        },
        {
          "fq_name": [
            "juniper-mx80-48t"
          ],
          "name": "juniper-mx80-48t"
        },
        {
          "fq_name": [
            "juniper-mx204"
          ],
          "name": "juniper-mx204"
        },
        {
          "fq_name": [
            "juniper-mx240"
          ],
          "name": "juniper-mx240"
        },
        {
          "fq_name": [
            "juniper-mx480"
          ],
          "name": "juniper-mx480"
        },
        {
          "fq_name": [
            "juniper-mx960"
          ],
          "name": "juniper-mx960"
        },
        {
          "fq_name": [
            "juniper-mx2008"
          ],
          "name": "juniper-mx2008"
        },
        {
          "fq_name": [
            "juniper-mx2010"
          ],
          "name": "juniper-mx2010"
        },
        {
          "fq_name": [
            "juniper-mx2020"
          ],
          "name": "juniper-mx2020"
        },
        {
          "fq_name": [
            "juniper-mx10003"
          ],
          "name": "juniper-mx10003"
        },
        {
          "fq_name": [
            "juniper-jnp10008"
          ],
          "name": "juniper-jnp10008"
        },
        {
          "fq_name": [
            "juniper-jnp10016"
          ],
          "name": "juniper-jnp10016"
        },
        {
          "fq_name": [
            "juniper-vmx"
          ],
          "name": "juniper-vmx"
        },
        {
          "fq_name": [
            "juniper-srx5400"
          ],
          "name": "juniper-srx5400"
        },
        {
          "fq_name": [
            "juniper-srx5600"
          ],
          "name": "juniper-srx5600"
        },
        {
          "fq_name": [
            "juniper-srx4600"
          ],
          "name": "juniper-srx4600"
        },
        {
          "fq_name": [
            "juniper-srx4100"
          ],
          "name": "juniper-srx4100"
        },
        {
          "fq_name": [
            "juniper-srx1500"
          ],
          "name": "juniper-srx1500"
        },
        {
          "fq_name": [
            "juniper-srx240h-poe"
          ],
          "name": "juniper-srx240h-poe"
        },
        {
          "fq_name": [
            "juniper-srx5800"
          ],
          "name": "juniper-srx5800"
        },
        {
          "fq_name": [
            "juniper-srx4200"
          ],
          "name": "juniper-srx4200"
        },
        {
          "fq_name": [
            "juniper-vsrx"
          ],
          "name": "juniper-vsrx"
        }
      ]
    },
    {
      "object_type": "feature",
      "objects": [
        {
          "fq_name": [
            "default-global-system-config", "underlay-ip-clos"
          ],
          "name": "underlay-ip-clos"
        },
        {
          "fq_name": [
            "default-global-system-config", "overlay-bgp"
          ],
          "name": "overlay-bgp"
        },
        {
          "fq_name": [
            "default-global-system-config", "l2-gateway"
          ],
          "name": "l2-gateway"
        },
        {
          "fq_name": [
            "default-global-system-config", "l3-gateway"
          ],
          "name": "l3-gateway"
        },
        {
          "fq_name": [
            "default-global-system-config", "vn-interconnect"
          ],
          "name": "vn-interconnect"
        },
        {
          "fq_name": [
            "default-global-system-config", "assisted-replicator"
          ],
          "name": "assisted-replicator"
        },
        {
          "fq_name": [
            "default-global-system-config", "port-profile"
          ],
          "name": "port-profile"
        },
        {
          "fq_name": [
            "default-global-system-config", "telemetry"
          ],
          "name": "telemetry"
        },
        {
          "fq_name": [
            "default-global-system-config", "firewall"
          ],
          "name": "firewall"
        },
        {
          "fq_name": [
            "default-global-system-config", "infra-bms-access"
          ],
          "name": "infra-bms-access"
        }
      ]
    },
    {
      "object_type": "physical-role",
      "objects": [
        {
          "fq_name": [
            "default-global-system-config", "leaf"
          ],
          "name": "leaf"
        },
        {
          "fq_name": [
            "default-global-system-config", "spine"
          ],
          "name": "spine"
        }
      ]
    },
    {
      "object_type": "overlay-role",
      "objects": [
        {
          "fq_name": [
            "default-global-system-config", "erb-ucast-gateway"
          ],
          "name": "erb-ucast-gateway"
        },
        {
          "fq_name": [
            "default-global-system-config", "crb-access"
          ],
          "name": "crb-access"
        },
        {
          "fq_name": [
            "default-global-system-config", "crb-mcast-gateway"
          ],
          "name": "crb-mcast-gateway"
        },
        {
          "fq_name": [
            "default-global-system-config", "crb-gateway"
          ],
          "name": "crb-gateway"
        },
        {
          "fq_name": [
            "default-global-system-config", "collapsed-spine"
          ],
          "name": "collapsed-spine"
        },
        {
          "fq_name": [
            "default-global-system-config", "crb-mcast-gateway"
          ],
          "name": "crb-mcast-gateway"
        },
        {
          "fq_name": [
            "default-global-system-config", "dc-gateway"
          ],
          "name": "dc-gateway"
        },
        {
          "fq_name": [
            "default-global-system-config", "dci-gateway"
          ],
          "name": "dci-gateway"
        },
        {
          "fq_name": [
            "default-global-system-config", "ar-client"
          ],
          "name": "ar-client"
        },
        {
          "fq_name": [
            "default-global-system-config", "ar-replicator"
          ],
          "name": "ar-replicator"
        },
        {
          "fq_name": [
            "default-global-system-config", "route-reflector"
          ],
          "name": "route-reflector"
        },
        {
          "fq_name": [
            "default-global-system-config", "lean"
          ],
          "name": "lean"
        }

      ]
    },
    {
      "object_type": "role-definition",
      "objects": [
        {
          "fq_name": [
            "default-global-system-config", "erb-leaf"
          ],
          "name": "erb-leaf"
        },
        {
          "fq_name": [
            "default-global-system-config", "crb-access-leaf"
          ],
          "name": "crb-access-leaf"
        },
        {
          "fq_name": [
            "default-global-system-config", "crb-access-spine"
          ],
          "name": "crb-access-spine"
        },
        {
          "fq_name": [
            "default-global-system-config", "crb-gateway-spine"
          ],
          "name": "crb-gateway-spine"
        },
        {
          "fq_name": [
            "default-global-system-config", "crb-mcast-gateway-spine"
          ],
          "name": "crb-mcast-gateway-spine"
        },
        {
          "fq_name": [
            "default-global-system-config", "collapsed-spine"
          ],
          "name": "collapsed-spine"
        },
        {
          "fq_name": [
            "default-global-system-config", "dc-gateway-spine"
          ],
          "name": "dc-gateway-spine"
        },
        {
          "fq_name": [
            "default-global-system-config", "dci-gateway-spine"
          ],
          "name": "dci-gateway-spine"
        },
        {
          "fq_name": [
            "default-global-system-config", "crb-gateway-leaf"
          ],
          "name": "crb-gateway-leaf"
        },
        {
          "fq_name": [
            "default-global-system-config", "crb-mcast-gateway-leaf"
          ],
          "name": "crb-mcast-gateway-leaf"
        },
        {
          "fq_name": [
            "default-global-system-config", "dc-gateway-leaf"
          ],
          "name": "dc-gateway-leaf"
        },
        {
          "fq_name": [
            "default-global-system-config", "dci-gateway-leaf"
          ],
          "name": "dci-gateway-leaf"
        },
        {
          "fq_name": [
            "default-global-system-config", "ar-client-leaf"
          ],
          "name": "ar-client-leaf"
        },
        {
          "fq_name": [
            "default-global-system-config", "ar-client-spine"
          ],
          "name": "ar-client-spine"
        },
        {
          "fq_name": [
            "default-global-system-config", "ar-replicator-leaf"
          ],
          "name": "ar-replicator-leaf"
        },
        {
          "fq_name": [
            "default-global-system-config", "ar-replicator-spine"
          ],
          "name": "ar-replicator-spine"
        },
        {
          "fq_name": [
            "default-global-system-config", "rr-spine"
          ],
          "name": "rr-spine"
        },
        {
          "fq_name": [
            "default-global-system-config", "rr-leaf"
          ],
          "name": "rr-leaf"
        },
        {
          "fq_name": [
            "default-global-system-config", "lean-spine"
          ],
          "name": "lean-spine"
        }
      ]
    },
    {
      "object_type": "feature-config",
      "objects": [
        {
          "fq_name": [
            "default-global-system-config", "erb-leaf", "l3-gateway"
          ],
          "name": "l3-gateway",
          "parent_type": "role-definition",
          "feature_config_additional_params": {
            "key_value_pair": [
              {
                "key": "use_gateway_ip",
                "value": "True"
              }
            ]
          }
        },
        {
          "fq_name": [
            "default-global-system-config", "ar-client-spine", "assisted-replicator"
          ],
          "name": "assisted-replicator",
          "parent_type": "role-definition",
          "feature_config_additional_params": {
            "key_value_pair": [
              {
                "key": "replicator_activation_delay",
                "value": "30"
              }
            ]
          }
        },
        {
          "fq_name": [
            "default-global-system-config", "ar-client-leaf", "assisted-replicator"
          ],
          "name": "assisted-replicator",
          "parent_type": "role-definition",
          "feature_config_additional_params": {
            "key_value_pair": [
              {
                "key": "replicator_activation_delay",
                "value": "30"
              }
            ]
          }
        }
      ]
    },
    {
      "object_type": "node-profile",
      "objects": [
        {
          "fq_name": [
            "default-global-system-config", "juniper-qfx5k-lean"
          ],
          "name": "juniper-qfx5k-lean",
          "node_profile_vendor": "Juniper",
          "node_profile_device_family": "junos-qfx",
          "node_profile_hitless_upgrade": true,
          "node_profile_roles": {
            "role_mappings": [
              {
                "physical_role": "leaf",
                "rb_roles": ["CRB-Access", "Route-Reflector", "AR-Client"]
              },
              {
                "physical_role": "spine",
                "rb_roles": ["lean", "Route-Reflector", "AR-Client"]
              }
            ]
          }
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-qfx5k"
          ],
          "name": "juniper-qfx5k",
          "node_profile_vendor": "Juniper",
          "node_profile_device_family": "junos-qfx",
          "node_profile_hitless_upgrade": true,
          "node_profile_roles": {
            "role_mappings": [
              {
                "physical_role": "leaf",
                "rb_roles": ["CRB-Access", "CRB-Gateway", "DC-Gateway", "Route-Reflector", "ERB-UCAST-Gateway", "DCI-Gateway", "PNF-Servicechain", "AR-Client"]
              },
              {
                "physical_role": "spine",
                "rb_roles": ["lean", "CRB-Access", "CRB-Gateway", "DC-Gateway", "Route-Reflector", "DCI-Gateway", "PNF-Servicechain", "AR-Client"]
              }
            ]
          }
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-qfx10k-lean"
          ],
          "name": "juniper-qfx10k-lean",
          "node_profile_vendor": "Juniper",
          "node_profile_device_family": "junos-qfx",
          "node_profile_hitless_upgrade": true,
          "node_profile_roles": {
            "role_mappings": [
              {
                "physical_role": "leaf",
                "rb_roles": ["Route-Reflector", "AR-Client"]
              },
              {
                "physical_role": "spine",
                "rb_roles": ["lean", "Route-Reflector", "AR-Client", "AR-Replicator"]
              }
            ]
          }
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-qfx5120"
          ],
          "name": "juniper-qfx5120",
          "node_profile_vendor": "Juniper",
          "node_profile_device_family": "junos-qfx",
          "node_profile_hitless_upgrade": true,
          "node_profile_roles": {
            "role_mappings": [
              {
                "physical_role": "leaf",
                "rb_roles": ["CRB-Access","Route-Reflector","AR-Client","ERB-UCAST-Gateway", "CRB-Gateway", "PNF-Servicechain"]
              },
              {
                "physical_role": "spine",
                "rb_roles": ["lean", "CRB-Gateway", "Route-Reflector", "PNF-Servicechain", "Collapsed-Spine"]
              }
            ]
          }
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-qfx10k"
          ],
          "name": "juniper-qfx10k",
          "node_profile_vendor": "Juniper",
          "node_profile_device_family": "junos-qfx",
          "node_profile_hitless_upgrade": true,
          "node_profile_roles": {
            "role_mappings": [
              {
                "physical_role": "leaf",
                "rb_roles": ["CRB-Access", "CRB-Gateway", "DC-Gateway", "Route-Reflector", "ERB-UCAST-Gateway", "DCI-Gateway", "CRB-MCAST-Gateway", "PNF-Servicechain", "AR-Client", "AR-Replicator"]
              },
              {
                "physical_role": "spine",
                "rb_roles": ["lean", "CRB-Access", "CRB-Gateway", "DC-Gateway", "Route-Reflector", "CRB-MCAST-Gateway", "DCI-Gateway", "PNF-Servicechain", "AR-Client", "AR-Replicator", "Collapsed-Spine"]
              }
            ]
          }
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-mx"
          ],
          "name": "juniper-mx",
          "node_profile_vendor": "Juniper",
          "node_profile_device_family": "junos",
          "node_profile_hitless_upgrade": true,
          "node_profile_roles": {
            "role_mappings": [
              {
                "physical_role": "leaf",
                "rb_roles": ["CRB-Gateway", "DC-Gateway", "Route-Reflector", "DCI-Gateway", "ERB-UCAST-Gateway", "DCI-Gateway", "CRB-MCAST-Gateway"]
              },
              {
                "physical_role": "spine",
                "rb_roles": ["lean", "CRB-Gateway", "DC-Gateway", "Route-Reflector", "CRB-MCAST-Gateway", "DCI-Gateway"]
              }
            ]
          }
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-srx"
          ],
          "name": "juniper-srx",
          "node_profile_vendor": "Juniper",
          "node_profile_device_family": "junos",
          "node_profile_hitless_upgrade": false,
          "node_profile_roles": {
            "role_mappings": [
              {
                "physical_role": "pnf",
                "rb_roles": ["PNF-Servicechain"]
              }
            ]
          }
        },
        {
          "fq_name": [
            "default-global-system-config", "device-functional-group"
          ],
          "name": "device-functional-group",
          "node_profile_vendor": "Juniper",
          "node_profile_roles": {
            "role_mappings": [
              {
                "physical_role": "leaf",
                "rb_roles": ["CRB-Access", "CRB-Gateway", "DC-Gateway", "Route-Reflector", "DCI-Gateway", "ERB-UCAST-Gateway", "CRB-MCAST-Gateway", "PNF-Servicechain", "AR-Client", "AR-Replicator"]
              },
              {
                "physical_role": "spine",
                "rb_roles": ["lean", "CRB-Access", "CRB-Gateway","DC-Gateway", "Route-Reflector", "CRB-MCAST-Gateway", "DCI-Gateway","ERB-UCAST-Gateway", "PNF-Servicechain", "AR-Client","AR-Replicator"]
              }
            ]
          }
        }
      ]
    },
    {
      "object_type": "telemetry-profile",
      "objects": [
        {
          "fq_name": [
            "default-domain", "default-project", "default-telemetry-profile-1"
          ],
          "name": "default-telemetry-profile-1",
          "telemetry_profile_is_default": true
        },
        {
          "fq_name": [
            "default-domain", "default-project", "default-telemetry-profile-2"
          ],
          "name": "default-telemetry-profile-2",
          "telemetry_profile_is_default": true
        },
        {
          "fq_name": [
            "default-domain", "default-project", "default-telemetry-profile-3"
          ],
          "name": "default-telemetry-profile-3",
          "telemetry_profile_is_default": true
        }
      ]
    },
    {
      "object_type": "sflow-profile",
      "objects": [
        {
          "fq_name": [
            "default-domain", "default-project", "sflow-all-interfaces"
          ],
          "name": "sflow-all-interfaces",
          "sflow_profile_is_default": true,
          "sflow_parameters": {
            "enabled_interface_type": "all",
            "stats_collection_frequency": {}
          }
        },
        {
          "fq_name": [
            "default-domain", "default-project", "sflow-fabric-interfaces"
          ],
          "name": "sflow-fabric-interfaces",
          "sflow_profile_is_default": true,
          "sflow_parameters": {
            "enabled_interface_type": "fabric",
            "stats_collection_frequency": {}
          }
        },
        {
          "fq_name": [
            "default-domain", "default-project", "sflow-access-interfaces"
          ],
          "name": "sflow-revenue-interfaces",
          "sflow_profile_is_default": true,
          "sflow_parameters": {
            "enabled_interface_type": "access",
            "stats_collection_frequency": {}
          }
        }
      ]
    },
    {
      "object_type": "device-functional-group",
      "objects": [
        {
          "fq_name": [
            "default-domain","default-project", "L2-Server-Leaf"
          ],
          "name": "L2-Server-Leaf",
          "device_functional_group_description": "Provides L2 servers connectivity with ingress replication for multicast in the spine",
          "device_functional_group_os_version": "18.4R2",
          "device_functional_group_routing_bridging_roles": {"rb_roles": ["CRB-Access","AR-Client"]}
        },
        {
          "fq_name": [
            "default-domain", "default-project", "L3-Server-Leaf"
          ],
          "name": "L3-Server-Leaf",
          "device_functional_group_description": "Provides L3 servers connectivity",
          "device_functional_group_os_version": "19.1R3",
          "device_functional_group_routing_bridging_roles": {"rb_roles": ["ERB-UCAST-Gateway"]}
        },
        {
          "fq_name": [
            "default-domain","default-project", "L3-Gateway-Spine"
          ],
          "name": "L3-Gateway-Spine",
          "device_functional_group_description": "Provides L3 gateway connectivity at spines",
          "device_functional_group_os_version": "18.4R2",
          "device_functional_group_routing_bridging_roles": {"rb_roles":["CRB-Gateway"]}

        },
        {
          "fq_name": [
            "default-domain","default-project", "L3-Storage-Leaf"
          ],
          "name": "L2-Storage-Leaf",
          "device_functional_group_description": "Provides L3 connectivity to storage arrays",
          "device_functional_group_os_version": "18.4R2",
          "device_functional_group_routing_bridging_roles": {"rb_roles":["ERB-UCAST-Gateway"]}

        },
         {
          "fq_name": [
            "default-domain","default-project", "L3-Server-Leaf-with-Optimized-Multicast"
          ],
          "name": "L3-Server-Leaf-with-Optimized-Multicast",
          "device_functional_group_description": "Provides L3 servers connectivity with Optimized multicast traffic",
          "device_functional_group_os_version": "18.4R2",
          "device_functional_group_routing_bridging_roles": {"rb_roles":["ERB-UCAST-Gateway","AR-Client"]}

        },
        {
          "fq_name": [
            "default-domain","default-project", "Centrally-Routed-Border-Spine"
          ],
          "name": "Centrally-Routed-Border-Spine",
          "device_functional_group_description": "Provides  L3 routing  for L2 server leafs and route reflector and ingress replication, as well DCGW and DCI GW and connectivity to firewalls",
          "device_functional_group_os_version": "18.4R2",
          "device_functional_group_routing_bridging_roles": {"rb_roles":["Route-Reflector","CRB-Gateway","DC-Gateway", "DCI-Gateway","PNF-Servicechain"]}

        },
        {
          "fq_name": [
            "default-domain","default-project", "Centrally-Routed-Border-Spine-With-Optimized-Multicast"
          ],
          "name": "Centrally Routed-Border-Spine-With-Optimized-Multicast",
          "device_functional_group_description": "Provides  L3 routing and gateway services for L2 server leafs ; provides route reflector and assisted replication",
          "device_functional_group_os_version": "18.4R2",
          "device_functional_group_routing_bridging_roles": {"rb_roles":["Route-Reflector","AR-Replicator", "CRB-Gateway","DC-Gateway","DCI-Gateway","PNF-Servicechain"]}

        },
        {
          "fq_name": [
            "default-domain","default-project", "Border-Spine-in-Edge-Routed"
          ],
          "name": "Border-Spine-in-Edge-Routed",
          "device_functional_group_description": "Provides L3 gateway services, route reflector",
          "device_functional_group_os_version": "18.4R2",
          "device_functional_group_routing_bridging_roles": {"rb_roles":["Route-Reflector","DC-Gateway","DCI-Gateway","PNF-Servicechain","ERB-UCAST-Gateway","CRB-MCAST-Gateway"]}

        },
        {
          "fq_name": [
            "default-domain","default-project", "Border-Leaf-in-Edge-Routed"
          ],
          "name": "Border-Leaf-in-Edge-Routed",
          "device_functional_group_description": "Provides L3 gateway services, route reflector",
          "device_functional_group_os_version": "18.4R2",
          "device_functional_group_routing_bridging_roles": {"rb_roles":["Route-Reflector","DC-Gateway","DCI-Gateway","PNF-Servicechain","ERB-UCAST-Gateway","CRB-MCAST-Gateway"]}

        },
         {
          "fq_name": [
            "default-domain","default-project", "Lean-Spine-with-Route-Reflector"
          ],
          "name": "Lean-Spine-with-Route-Reflector",
          "device_functional_group_description": "Spine only acting as Route Reflector",
          "device_functional_group_os_version": "18.4R2",
          "device_functional_group_routing_bridging_roles": {"rb_roles":["Route-Reflector","lean"]}

        }
      ]
    },
    {
      "object_type": "role-config",
      "objects": [
        {
          "fq_name": [
            "default-global-system-config", "juniper-qfx5k-lean", "basic"
          ],
          "name": "basic",
          "parent_type": "node-profile",
          "role_config_config": "{\"snmp\": {\"communities\": [{\"readonly\": true, \"name\": \"public\"}]}}"
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-qfx5k", "basic"
          ],
          "name": "basic",
          "parent_type": "node-profile",
          "role_config_config": "{\"snmp\": {\"communities\": [{\"readonly\": true, \"name\": \"public\"}]}}"
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-qfx5120", "basic"
          ],
          "name": "basic",
          "parent_type": "node-profile",
          "role_config_config": "{\"snmp\": {\"communities\": [{\"readonly\": true, \"name\": \"public\"}]}}"
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-qfx10k-lean", "basic"
          ],
          "name": "basic",
          "parent_type": "node-profile",
          "role_config_config": "{\"snmp\": {\"communities\": [{\"readonly\": true, \"name\": \"public\"}]}}"
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-qfx10k", "basic"
          ],
          "name": "basic",
          "parent_type": "node-profile",
          "role_config_config": "{\"snmp\": {\"communities\": [{\"readonly\": true, \"name\": \"public\"}]}}"
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-mx", "basic"
          ],
          "name": "basic",
          "parent_type": "node-profile",
          "role_config_config": "{\"snmp\": {\"communities\": [{\"readonly\": true, \"name\": \"public\"}]}}"
        },
        {
          "fq_name": [
            "default-global-system-config", "juniper-srx", "basic"
          ],
          "name": "basic",
          "parent_type": "node-profile",
          "role_config_config": "{\"snmp\": {\"communities\": [{\"readonly\": true, \"name\": \"public\"}]}}"
        }
      ]
    }
  ],
  "refs": [
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "erb-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "erb-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "erb-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l2-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "erb-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "firewall" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "erb-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l3-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "erb-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "vn-interconnect" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "erb-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "port-profile"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "erb-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "erb-leaf" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "erb-leaf" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "erb-ucast-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "port-profile"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l2-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "firewall" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-leaf" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-leaf" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "crb-access" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "port-profile"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l2-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "firewall" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-spine" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-spine" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "crb-access" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l3-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "vn-interconnect" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-spine" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "crb-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-spine" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l3-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "vn-interconnect" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-leaf" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "crb-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-leaf" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l3-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "vn-interconnect" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-spine" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "crb-mcast-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-spine" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l3-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "vn-interconnect" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-leaf" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "crb-mcast-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-leaf" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l3-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "vn-interconnect" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-spine" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "dc-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-spine" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "lean-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "lean-spine" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "lean" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "lean-spine" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l3-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "vn-interconnect" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-leaf" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "dc-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-leaf" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l3-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l2-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "vn-interconnect" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-spine" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "dci-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-spine" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l3-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l2-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "vn-interconnect" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-leaf" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "dci-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-leaf" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-spine" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "crb-access" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "assisted-replicator" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "assisted-replicator" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-client-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "assisted-replicator" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-client-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-client-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "assisted-replicator" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-client-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-leaf" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-leaf" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "ar-replicator" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-spine" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-spine" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "ar-replicator" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-client-leaf" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-client-leaf" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "ar-client" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-client-spine" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-spine" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-leaf" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-client-spine" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "ar-client" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-spine" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "route-reflector" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-leaf" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "route-reflector" ]
    },
        {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "erb-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-access-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dc-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "dci-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-client-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-client-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "ar-replicator-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "rr-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "lean-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l2-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "firewall" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l2-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "firewall" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l2-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "firewall" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l2-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "crb-mcast-gateway-leaf" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "firewall" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "lean-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "collapsed-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l3-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "collapsed-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "underlay-ip-clos" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "collapsed-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "vn-interconnect" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "collapsed-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "telemetry" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "collapsed-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "overlay-bgp" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "collapsed-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "l2-gateway" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "collapsed-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "firewall" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "collapsed-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "infra-bms-access"]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "collapsed-spine" ],
      "to_type": "feature",
      "to_fq_name": [ "default-global-system-config", "port-profile" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "collapsed-spine" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "collapsed-spine" ]
    },
    {
      "from_type": "role-definition",
      "from_fq_name": [ "default-global-system-config", "collapsed-spine" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100-48sh-afi" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100-48sh-afo" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100-48th-afi" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100-48th-afo" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100-48s-6q" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100-96s-8q" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100-48t" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100-48t-6q" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100-24q-2p" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100-24q-aa" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100e-48s-6q" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100e-96s-8q" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100e-48t-6q" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5100e-24q-2p" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5110-48s-4c" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5110-32q" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5120" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5120-48y-8c" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5120" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5120-32c" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5120" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5120-48t-6c" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5200-32c-32q" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5200-48y" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5210-64c" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5220-32cd" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx5220-128c" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx10k" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx10002-72q" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx10k" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx10002-36q" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx10k" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx10002-60c" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx10k" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx10016" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx10k" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-qfx10008" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx10k" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-vqfx-10000" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-mx80" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-mx80-48t" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-mx204" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-mx240" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-mx480" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-mx960" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-mx2008" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-mx2010" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-mx2020" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-mx10003" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-jnp10008" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-jnp10016" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-vmx" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-srx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-srx5400" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-srx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-srx5600" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-srx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-srx4600" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-srx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-srx4100" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-srx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-srx1500" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-srx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-srx240h-poe" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-srx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-vsrx" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-srx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-srx5800" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-srx" ],
      "to_type": "hardware",
      "to_fq_name": [ "juniper-srx4200" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k" ],
      "to_type": "physical-role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "erb-ucast-gateway" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "overlay-role",
      "to_fq_name": [ "default-global-system-config", "erb-ucast-gateway" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k-lean" ],
      "to_type": "job_template",
      "to_fq_name": [ "default-global-system-config", "fabric_config_template" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5k" ],
      "to_type": "job_template",
      "to_fq_name": [ "default-global-system-config", "fabric_config_template" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx5120" ],
      "to_type": "job_template",
      "to_fq_name": [ "default-global-system-config", "fabric_config_template" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx10k-lean" ],
      "to_type": "job_template",
      "to_fq_name": [ "default-global-system-config", "fabric_config_template" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-qfx10k" ],
      "to_type": "job_template",
      "to_fq_name": [ "default-global-system-config", "fabric_config_template" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-mx" ],
      "to_type": "job_template",
      "to_fq_name": [ "default-global-system-config", "fabric_config_template" ]
    },
    {
      "from_type": "node-profile",
      "from_fq_name": [ "default-global-system-config", "juniper-srx" ],
      "to_type": "job_template",
      "to_fq_name": [ "default-global-system-config", "fabric_config_template" ]
    },
    {
      "from_type": "telemetry-profile",
      "from_fq_name": [ "default-domain", "default-project", "default-telemetry-profile-1" ],
      "to_type": "sflow-profile",
      "to_fq_name": [ "default-domain", "default-project", "sflow-all-interfaces" ]
    },
    {
      "from_type": "telemetry-profile",
      "from_fq_name": [ "default-domain", "default-project", "default-telemetry-profile-2" ],
      "to_type": "sflow-profile",
      "to_fq_name": [ "default-domain", "default-project", "sflow-fabric-interfaces" ]
    },
    {
      "from_type": "telemetry-profile",
      "from_fq_name": [ "default-domain", "default-project", "default-telemetry-profile-3" ],
      "to_type": "sflow-profile",
      "to_fq_name": [ "default-domain", "default-project", "sflow-access-interfaces" ]
    },
    {
      "from_type": "device-functional-group",
      "from_fq_name": [ "default-domain", "default-project","L2-Server-Leaf" ],
      "to_type": "physical_role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "device-functional-group",
      "from_fq_name": [ "default-domain", "default-project","L3-Storage-Leaf" ],
      "to_type": "physical_role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "device-functional-group",
      "from_fq_name": [ "default-domain", "default-project","L3-Gateway-Spine" ],
      "to_type": "physical_role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "device-functional-group",
      "from_fq_name": [ "default-domain", "default-project","L3-Server-Leaf-with-Optimized-Multicast" ],
      "to_type": "physical_role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "device-functional-group",
      "from_fq_name": [ "default-domain", "default-project","Centrally-Routed-Border-Spine" ],
      "to_type": "physical_role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "device-functional-group",
      "from_fq_name": [ "default-domain", "default-project","Centrally-Routed-Border-Spine-With-Optimized-Multicast" ],
      "to_type": "physical_role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "device-functional-group",
      "from_fq_name": [ "default-domain", "default-project","Border-Spine-in-Edge-Routed" ],
      "to_type": "physical_role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    },
    {
      "from_type": "device-functional-group",
      "from_fq_name": [ "default-domain", "default-project","Border-Leaf-in-Edge-Routed" ],
      "to_type": "physical_role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "device-functional-group",
      "from_fq_name": [ "default-domain", "default-project","L3-Server-Leaf" ],
      "to_type": "physical_role",
      "to_fq_name": [ "default-global-system-config", "leaf" ]
    },
    {
      "from_type": "device-functional-group",
      "from_fq_name": [ "default-domain", "default-project","Lean-Spine-with-Route-Reflector" ],
      "to_type": "physical_role",
      "to_fq_name": [ "default-global-system-config", "spine" ]
    }
  ]
}
`
