package k8s

// Container represents a single container in a Pod (core/v1.Container).
// All nested Kubernetes types are fully dereferenced to primitives.
type Container struct {
	Name                     string               `json:"name"`
	Image                    string               `json:"image,omitempty"`
	Command                  []string             `json:"command,omitempty"`
	Args                     []string             `json:"args,omitempty"`
	WorkingDir               string               `json:"workingDir,omitempty"`
	Ports                    []ContainerPort      `json:"ports,omitempty"`
	EnvFrom                  []EnvFromSource      `json:"envFrom,omitempty"`
	Env                      []EnvVar             `json:"env,omitempty"`
	Resources                ResourceRequirements `json:"resources,omitempty"`
	VolumeMounts             []VolumeMount        `json:"volumeMounts,omitempty"`
	VolumeDevices            []VolumeDevice       `json:"volumeDevices,omitempty"`
	LivenessProbe            *Probe               `json:"livenessProbe,omitempty"`
	ReadinessProbe           *Probe               `json:"readinessProbe,omitempty"`
	StartupProbe             *Probe               `json:"startupProbe,omitempty"`
	Lifecycle                *Lifecycle           `json:"lifecycle,omitempty"`
	TerminationMessagePath   string               `json:"terminationMessagePath,omitempty"`
	TerminationMessagePolicy string               `json:"terminationMessagePolicy,omitempty"`
	ImagePullPolicy          string               `json:"imagePullPolicy,omitempty"`
	SecurityContext          *SecurityContext     `json:"securityContext,omitempty"`
	Stdin                    bool                 `json:"stdin,omitempty"`
	StdinOnce                bool                 `json:"stdinOnce,omitempty"`
	TTY                      bool                 `json:"tty,omitempty"`
}

// Supporting types (all inlined, no external imports needed).

type ContainerPort struct {
	Name          string `json:"name,omitempty"`
	HostPort      int32  `json:"hostPort,omitempty"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"` // Default "TCP"
	HostIP        string `json:"hostIP,omitempty"`
}

type EnvFromSource struct {
	Prefix       string              `json:"prefix,omitempty"`
	ConfigMapRef *ConfigMapEnvSource `json:"configMapRef,omitempty"`
	SecretRef    *SecretEnvSource    `json:"secretRef,omitempty"`
}

type ConfigMapEnvSource struct {
	LocalObjectReference LocalObjectReference `json:"localObjectReference"`
	Optional             *bool                `json:"optional,omitempty"`
}

type SecretEnvSource struct {
	LocalObjectReference LocalObjectReference `json:"localObjectReference"`
	Optional             *bool                `json:"optional,omitempty"`
}

type EnvVar struct {
	Name      string        `json:"name"`
	Value     string        `json:"value,omitempty"`
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty"`
}

type EnvVarSource struct {
	FieldRef         *ObjectFieldSelector   `json:"fieldRef,omitempty"`
	ResourceFieldRef *ResourceFieldSelector `json:"resourceFieldRef,omitempty"`
	ConfigMapKeyRef  *ConfigMapKeySelector  `json:"configMapKeyRef,omitempty"`
	SecretKeyRef     *SecretKeySelector     `json:"secretKeyRef,omitempty"`
}

type ObjectFieldSelector struct {
	APIVersion string `json:"apiVersion,omitempty"` // default "v1"
	FieldPath  string `json:"fieldPath"`
}

type ResourceFieldSelector struct {
	ContainerName string `json:"containerName,omitempty"`
	Resource      string `json:"resource"`
	Divisor       string `json:"divisor,omitempty"` // Quantity as string
}

type ConfigMapKeySelector struct {
	LocalObjectReference LocalObjectReference `json:"localObjectReference"`
	Key                  string               `json:"key"`
	Optional             *bool                `json:"optional,omitempty"`
}

type SecretKeySelector struct {
	LocalObjectReference LocalObjectReference `json:"localObjectReference"`
	Key                  string               `json:"key"`
	Optional             *bool                `json:"optional,omitempty"`
}

type LocalObjectReference struct {
	Name string `json:"name,omitempty"`
}

type ResourceRequirements struct {
	Limits   map[string]string `json:"limits,omitempty"`   // e.g. "100m", "500Mi" as string
	Requests map[string]string `json:"requests,omitempty"` // e.g. "100m", "500Mi" as string
}

type VolumeMount struct {
	Name             string  `json:"name"`
	ReadOnly         bool    `json:"readOnly,omitempty"`
	MountPath        string  `json:"mountPath"`
	SubPath          string  `json:"subPath,omitempty"`
	MountPropagation *string `json:"mountPropagation,omitempty"`
	SubPathExpr      string  `json:"subPathExpr,omitempty"`
}

type VolumeDevice struct {
	Name       string `json:"name"`
	DevicePath string `json:"devicePath"`
}

type Probe struct {
	// ProbeHandler fields inlined for direct compilation without embedding.
	Exec                          *ExecAction      `json:"exec,omitempty"`
	HTTPGet                       *HTTPGetAction   `json:"httpGet,omitempty"`
	TCPSocket                     *TCPSocketAction `json:"tcpSocket,omitempty"`
	GRPC                          *GRPCAction      `json:"grpc,omitempty"`
	InitialDelaySeconds           int32            `json:"initialDelaySeconds,omitempty"`
	TimeoutSeconds                int32            `json:"timeoutSeconds,omitempty"`
	PeriodSeconds                 int32            `json:"periodSeconds,omitempty"`
	SuccessThreshold              int32            `json:"successThreshold,omitempty"`
	FailureThreshold              int32            `json:"failureThreshold,omitempty"`
	TerminationGracePeriodSeconds *int64           `json:"terminationGracePeriodSeconds,omitempty"`
}

type ExecAction struct {
	Command []string `json:"command,omitempty"`
}

type HTTPGetAction struct {
	Path        string       `json:"path,omitempty"`
	Port        int32        `json:"port"`
	Host        string       `json:"host,omitempty"`
	Scheme      string       `json:"scheme,omitempty"`
	HTTPHeaders []HTTPHeader `json:"httpHeaders,omitempty"`
}

type HTTPHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type TCPSocketAction struct {
	Port int32  `json:"port"`
	Host string `json:"host,omitempty"`
}

type GRPCAction struct {
	Port    int32   `json:"port"`
	Service *string `json:"service,omitempty"`
}

type Lifecycle struct {
	PostStart *LifecycleHandler `json:"postStart,omitempty"`
	PreStop   *LifecycleHandler `json:"preStop,omitempty"`
}

type LifecycleHandler struct {
	Exec      *ExecAction      `json:"exec,omitempty"`
	HTTPGet   *HTTPGetAction   `json:"httpGet,omitempty"`
	TCPSocket *TCPSocketAction `json:"tcpSocket,omitempty"`
}

type SecurityContext struct {
	Capabilities             *Capabilities                  `json:"capabilities,omitempty"`
	Privileged               *bool                          `json:"privileged,omitempty"`
	SELinuxOptions           *SELinuxOptions                `json:"seLinuxOptions,omitempty"`
	WindowsOptions           *WindowsSecurityContextOptions `json:"windowsOptions,omitempty"`
	RunAsUser                *int64                         `json:"runAsUser,omitempty"`
	RunAsGroup               *int64                         `json:"runAsGroup,omitempty"`
	RunAsNonRoute            *bool                          `json:"runAsNonRoute,omitempty"`
	ReadOnlyRouteFilesystem  *bool                          `json:"readOnlyRouteFilesystem,omitempty"`
	AllowPrivilegeEscalation *bool                          `json:"allowPrivilegeEscalation,omitempty"`
	ProcMount                *string                        `json:"procMount,omitempty"`
	SeccompProfile           *SeccompProfile                `json:"seccompProfile,omitempty"`
}

type Capabilities struct {
	Add  []string `json:"add,omitempty"`
	Drop []string `json:"drop,omitempty"`
}

type SELinuxOptions struct {
	User  string `json:"user,omitempty"`
	Role  string `json:"role,omitempty"`
	Type  string `json:"type,omitempty"`
	Level string `json:"level,omitempty"`
}

type WindowsSecurityContextOptions struct {
	GMSACredentialSpecName *string `json:"gmsaCredentialSpecName,omitempty"`
	GMSACredentialSpec     *string `json:"gmsaCredentialSpec,omitempty"`
	RunAsUserName          *string `json:"runAsUserName,omitempty"`
	HostProcess            *bool   `json:"hostProcess,omitempty"`
}

type SeccompProfile struct {
	Type             string  `json:"type"`
	LocalhostProfile *string `json:"localhostProfile,omitempty"`
}
