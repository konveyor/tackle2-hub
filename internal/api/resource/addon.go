package resource

import (
	"encoding/json"

	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/api/k8s"
	core "k8s.io/api/core/v1"
)

// Addon REST resource.
type Addon api.Addon

// With model.
func (r *Addon) With(m *crd.Addon, extensions ...crd.Extension) {
	r.Name = m.Name
	r.Container = convertContainer(m.Spec.Container)
	if m.Spec.Metadata.Raw != nil {
		_ = json.Unmarshal(m.Spec.Metadata.Raw, &r.Metadata)
	}
	for i := range extensions {
		extension := Extension{}
		extension.With(&extensions[i])
		r.Extensions = append(
			r.Extensions,
			api.Extension(extension))
	}
}

// Extension REST resource.
type Extension api.Extension

// With model.
func (r *Extension) With(m *crd.Extension) {
	r.Name = m.Name
	r.Addon = m.Spec.Addon
	r.Container = convertContainer(m.Spec.Container)
	if m.Spec.Metadata.Raw != nil {
		_ = json.Unmarshal(m.Spec.Metadata.Raw, &r.Metadata)
	}
}

// convertContainer converts a k8s.io/api/core/v1.Container to k8s.Container.
func convertContainer(src core.Container) k8s.Container {
	dst := k8s.Container{
		Name:                     src.Name,
		Image:                    src.Image,
		Command:                  src.Command,
		Args:                     src.Args,
		WorkingDir:               src.WorkingDir,
		TerminationMessagePath:   src.TerminationMessagePath,
		TerminationMessagePolicy: string(src.TerminationMessagePolicy),
		ImagePullPolicy:          string(src.ImagePullPolicy),
		Stdin:                    src.Stdin,
		StdinOnce:                src.StdinOnce,
		TTY:                      src.TTY,
	}
	// Convert Ports
	for _, port := range src.Ports {
		dst.Ports = append(dst.Ports, k8s.ContainerPort{
			Name:          port.Name,
			HostPort:      port.HostPort,
			ContainerPort: port.ContainerPort,
			Protocol:      string(port.Protocol),
			HostIP:        port.HostIP,
		})
	}
	// Convert EnvFrom
	for _, envFrom := range src.EnvFrom {
		e := k8s.EnvFromSource{
			Prefix: envFrom.Prefix,
		}
		if envFrom.ConfigMapRef != nil {
			e.ConfigMapRef = &k8s.ConfigMapEnvSource{
				LocalObjectReference: k8s.LocalObjectReference{Name: envFrom.ConfigMapRef.Name},
				Optional:             envFrom.ConfigMapRef.Optional,
			}
		}
		if envFrom.SecretRef != nil {
			e.SecretRef = &k8s.SecretEnvSource{
				LocalObjectReference: k8s.LocalObjectReference{Name: envFrom.SecretRef.Name},
				Optional:             envFrom.SecretRef.Optional,
			}
		}
		dst.EnvFrom = append(dst.EnvFrom, e)
	}

	// Convert Env
	for _, env := range src.Env {
		e := k8s.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		}
		if env.ValueFrom != nil {
			e.ValueFrom = &k8s.EnvVarSource{}
			if env.ValueFrom.FieldRef != nil {
				e.ValueFrom.FieldRef = &k8s.ObjectFieldSelector{
					APIVersion: env.ValueFrom.FieldRef.APIVersion,
					FieldPath:  env.ValueFrom.FieldRef.FieldPath,
				}
			}
			if env.ValueFrom.ResourceFieldRef != nil {
				divisor := ""
				if !env.ValueFrom.ResourceFieldRef.Divisor.IsZero() {
					divisor = env.ValueFrom.ResourceFieldRef.Divisor.String()
				}
				e.ValueFrom.ResourceFieldRef = &k8s.ResourceFieldSelector{
					ContainerName: env.ValueFrom.ResourceFieldRef.ContainerName,
					Resource:      env.ValueFrom.ResourceFieldRef.Resource,
					Divisor:       divisor,
				}
			}
			if env.ValueFrom.ConfigMapKeyRef != nil {
				e.ValueFrom.ConfigMapKeyRef = &k8s.ConfigMapKeySelector{
					LocalObjectReference: k8s.LocalObjectReference{Name: env.ValueFrom.ConfigMapKeyRef.Name},
					Key:                  env.ValueFrom.ConfigMapKeyRef.Key,
					Optional:             env.ValueFrom.ConfigMapKeyRef.Optional,
				}
			}
			if env.ValueFrom.SecretKeyRef != nil {
				e.ValueFrom.SecretKeyRef = &k8s.SecretKeySelector{
					LocalObjectReference: k8s.LocalObjectReference{Name: env.ValueFrom.SecretKeyRef.Name},
					Key:                  env.ValueFrom.SecretKeyRef.Key,
					Optional:             env.ValueFrom.SecretKeyRef.Optional,
				}
			}
		}
		dst.Env = append(dst.Env, e)
	}
	// Convert Resources
	dst.Resources.Limits = make(map[string]string)
	dst.Resources.Requests = make(map[string]string)
	for k, v := range src.Resources.Limits {
		dst.Resources.Limits[string(k)] = v.String()
	}
	for k, v := range src.Resources.Requests {
		dst.Resources.Requests[string(k)] = v.String()
	}
	// Convert VolumeMounts
	for _, vm := range src.VolumeMounts {
		mount := k8s.VolumeMount{
			Name:        vm.Name,
			ReadOnly:    vm.ReadOnly,
			MountPath:   vm.MountPath,
			SubPath:     vm.SubPath,
			SubPathExpr: vm.SubPathExpr,
		}
		if vm.MountPropagation != nil {
			prop := string(*vm.MountPropagation)
			mount.MountPropagation = &prop
		}
		dst.VolumeMounts = append(dst.VolumeMounts, mount)
	}
	// Convert VolumeDevices
	for _, vd := range src.VolumeDevices {
		dst.VolumeDevices = append(dst.VolumeDevices, k8s.VolumeDevice{
			Name:       vd.Name,
			DevicePath: vd.DevicePath,
		})
	}
	// Convert Probes
	if src.LivenessProbe != nil {
		dst.LivenessProbe = convertProbe(src.LivenessProbe)
	}
	if src.ReadinessProbe != nil {
		dst.ReadinessProbe = convertProbe(src.ReadinessProbe)
	}
	if src.StartupProbe != nil {
		dst.StartupProbe = convertProbe(src.StartupProbe)
	}
	// Convert Lifecycle
	if src.Lifecycle != nil {
		dst.Lifecycle = &k8s.Lifecycle{}
		if src.Lifecycle.PostStart != nil {
			dst.Lifecycle.PostStart = convertLifecycleHandler(src.Lifecycle.PostStart)
		}
		if src.Lifecycle.PreStop != nil {
			dst.Lifecycle.PreStop = convertLifecycleHandler(src.Lifecycle.PreStop)
		}
	}
	// Convert SecurityContext
	if src.SecurityContext != nil {
		dst.SecurityContext = convertSecurityContext(src.SecurityContext)
	}
	return dst
}

// convertProbe converts a core.Probe to k8s.Probe.
func convertProbe(src *core.Probe) *k8s.Probe {
	dst := &k8s.Probe{
		InitialDelaySeconds:           src.InitialDelaySeconds,
		TimeoutSeconds:                src.TimeoutSeconds,
		PeriodSeconds:                 src.PeriodSeconds,
		SuccessThreshold:              src.SuccessThreshold,
		FailureThreshold:              src.FailureThreshold,
		TerminationGracePeriodSeconds: src.TerminationGracePeriodSeconds,
	}
	if src.Exec != nil {
		dst.Exec = &k8s.ExecAction{
			Command: src.Exec.Command,
		}
	}
	if src.HTTPGet != nil {
		httpGet := &k8s.HTTPGetAction{
			Path:   src.HTTPGet.Path,
			Port:   src.HTTPGet.Port.IntVal,
			Host:   src.HTTPGet.Host,
			Scheme: string(src.HTTPGet.Scheme),
		}
		for _, h := range src.HTTPGet.HTTPHeaders {
			httpGet.HTTPHeaders = append(httpGet.HTTPHeaders, k8s.HTTPHeader{
				Name:  h.Name,
				Value: h.Value,
			})
		}
		dst.HTTPGet = httpGet
	}
	if src.TCPSocket != nil {
		dst.TCPSocket = &k8s.TCPSocketAction{
			Port: src.TCPSocket.Port.IntVal,
			Host: src.TCPSocket.Host,
		}
	}
	if src.GRPC != nil {
		dst.GRPC = &k8s.GRPCAction{
			Port:    src.GRPC.Port,
			Service: src.GRPC.Service,
		}
	}
	return dst
}

// convertLifecycleHandler converts a core.LifecycleHandler to k8s.LifecycleHandler.
func convertLifecycleHandler(src *core.LifecycleHandler) *k8s.LifecycleHandler {
	dst := &k8s.LifecycleHandler{}
	if src.Exec != nil {
		dst.Exec = &k8s.ExecAction{
			Command: src.Exec.Command,
		}
	}
	if src.HTTPGet != nil {
		httpGet := &k8s.HTTPGetAction{
			Path:   src.HTTPGet.Path,
			Port:   src.HTTPGet.Port.IntVal,
			Host:   src.HTTPGet.Host,
			Scheme: string(src.HTTPGet.Scheme),
		}
		for _, h := range src.HTTPGet.HTTPHeaders {
			httpGet.HTTPHeaders = append(httpGet.HTTPHeaders, k8s.HTTPHeader{
				Name:  h.Name,
				Value: h.Value,
			})
		}
		dst.HTTPGet = httpGet
	}
	if src.TCPSocket != nil {
		dst.TCPSocket = &k8s.TCPSocketAction{
			Port: src.TCPSocket.Port.IntVal,
			Host: src.TCPSocket.Host,
		}
	}
	return dst
}

// convertSecurityContext converts a core.SecurityContext to k8s.SecurityContext.
func convertSecurityContext(src *core.SecurityContext) *k8s.SecurityContext {
	dst := &k8s.SecurityContext{
		Privileged:               src.Privileged,
		RunAsUser:                src.RunAsUser,
		RunAsGroup:               src.RunAsGroup,
		RunAsNonRoute:            src.RunAsNonRoot,
		ReadOnlyRouteFilesystem:  src.ReadOnlyRootFilesystem,
		AllowPrivilegeEscalation: src.AllowPrivilegeEscalation,
	}
	if src.Capabilities != nil {
		dst.Capabilities = &k8s.Capabilities{
			Add:  make([]string, len(src.Capabilities.Add)),
			Drop: make([]string, len(src.Capabilities.Drop)),
		}
		for i, cap := range src.Capabilities.Add {
			dst.Capabilities.Add[i] = string(cap)
		}
		for i, cap := range src.Capabilities.Drop {
			dst.Capabilities.Drop[i] = string(cap)
		}
	}
	if src.SELinuxOptions != nil {
		dst.SELinuxOptions = &k8s.SELinuxOptions{
			User:  src.SELinuxOptions.User,
			Role:  src.SELinuxOptions.Role,
			Type:  src.SELinuxOptions.Type,
			Level: src.SELinuxOptions.Level,
		}
	}
	if src.WindowsOptions != nil {
		dst.WindowsOptions = &k8s.WindowsSecurityContextOptions{
			GMSACredentialSpecName: src.WindowsOptions.GMSACredentialSpecName,
			GMSACredentialSpec:     src.WindowsOptions.GMSACredentialSpec,
			RunAsUserName:          src.WindowsOptions.RunAsUserName,
			HostProcess:            src.WindowsOptions.HostProcess,
		}
	}
	if src.ProcMount != nil {
		procMount := string(*src.ProcMount)
		dst.ProcMount = &procMount
	}
	if src.SeccompProfile != nil {
		dst.SeccompProfile = &k8s.SeccompProfile{
			Type:             string(src.SeccompProfile.Type),
			LocalhostProfile: src.SeccompProfile.LocalhostProfile,
		}
	}
	return dst
}
