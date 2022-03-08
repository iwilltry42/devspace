package loader

import (
	"bytes"
	"fmt"
	jsonyaml "github.com/ghodss/yaml"
	"github.com/loft-sh/devspace/pkg/devspace/config/versions/latest"
	"github.com/loft-sh/devspace/pkg/devspace/imageselector"
	"github.com/loft-sh/devspace/pkg/util/encoding"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	k8sv1 "k8s.io/api/core/v1"
)

// ValidInitialSyncStrategy checks if strategy is valid
func ValidInitialSyncStrategy(strategy latest.InitialSyncStrategy) bool {
	return strategy == "" ||
		strategy == latest.InitialSyncStrategyMirrorLocal ||
		strategy == latest.InitialSyncStrategyMirrorRemote ||
		strategy == latest.InitialSyncStrategyKeepAll ||
		strategy == latest.InitialSyncStrategyPreferLocal ||
		strategy == latest.InitialSyncStrategyPreferRemote ||
		strategy == latest.InitialSyncStrategyPreferNewest
}

// ValidContainerArch checks if the target container arch is valid
func ValidContainerArch(arch latest.ContainerArchitecture) bool {
	return arch == "" ||
		arch == latest.ContainerArchitectureAmd64 ||
		arch == latest.ContainerArchitectureArm64
}

func Validate(config *latest.Config) error {
	if config.Name == "" {
		return fmt.Errorf("you need to specify a name for your devspace.yaml")
	}
	if encoding.IsUnsafeName(config.Name) {
		return fmt.Errorf("name has to match the following regex: %v", encoding.UnsafeNameRegEx.String())
	}

	err := validateRequire(config)
	if err != nil {
		return err
	}

	err = validateImages(config)
	if err != nil {
		return err
	}

	err = validateDev(config)
	if err != nil {
		return err
	}

	err = validateHooks(config)
	if err != nil {
		return err
	}

	err = validateDeployments(config)
	if err != nil {
		return err
	}

	err = validatePullSecrets(config)
	if err != nil {
		return err
	}

	err = validateCommands(config)
	if err != nil {
		return err
	}

	err = validateDependencies(config)
	if err != nil {
		return err
	}

	return nil
}

func validateVars(vars []*latest.Variable) error {
	for i, v := range vars {
		if v.Name == "" {
			return fmt.Errorf("vars[*].name has to be specified")
		}
		if encoding.IsUnsafeUpperName(v.Name) {
			return fmt.Errorf("vars[%d].name %s has to match the following regex: %v", i, v.Name, encoding.UnsafeUpperNameRegEx.String())
		}

		// make sure is unique
		for j, v2 := range vars {
			if i != j && v.Name == v2.Name {
				return fmt.Errorf("multiple definitions for variable %s found", v.Name)
			}
		}
	}

	return nil
}

func validateRequire(config *latest.Config) error {
	for index, plugin := range config.Require.Plugins {
		if plugin.Name == "" {
			return errors.Errorf("require.plugins[%d].name is required", index)
		}
		if plugin.Version == "" {
			return errors.Errorf("require.plugins[%d].version is required", index)
		}
	}

	for index, command := range config.Require.Commands {
		if command.Name == "" {
			return errors.Errorf("require.commands[%d].name is required", index)
		}
		if command.Version == "" {
			return errors.Errorf("require.commands[%d].version is required", index)
		}
	}

	return nil
}

func validateDependencies(config *latest.Config) error {
	for name, dep := range config.Dependencies {
		if encoding.IsUnsafeName(name) {
			return fmt.Errorf("dependencies[%s] has to match the following regex: %v", name, encoding.UnsafeNameRegEx.String())
		}
		if dep.Source == nil {
			return errors.Errorf("dependencies[%s].source is required", name)
		}
		if dep.Source.Git == "" && dep.Source.Path == "" {
			return errors.Errorf("dependencies[%s].source.git or dependencies[%s].source.path is required", name, name)
		}
		if len(dep.Profiles) > 0 && dep.Profile != "" {
			return errors.Errorf("dependencies[%s].profiles and dependencies[%s].profile & dependencies[%s].profileParents cannot be used together", name, name, name)
		}
	}

	return nil
}

func validateCommands(config *latest.Config) error {
	for key, command := range config.Commands {
		if command.Name == "" {
			return errors.Errorf("commands[%s].name is required", key)
		}
		if encoding.IsUnsafeUpperName(command.Name) {
			return fmt.Errorf("commands[%s] has to match the following regex: %v", command.Name, encoding.UnsafeUpperNameRegEx.String())
		}
		if command.Command == "" {
			return errors.Errorf("commands[%s].command is required", key)
		}
	}

	return nil
}

func validateHooks(config *latest.Config) error {
	for index, hookConfig := range config.Hooks {
		if len(hookConfig.Events) == 0 {
			return errors.Errorf("hooks[%d].events is required", index)
		}
		if hookConfig.Command == "" && hookConfig.Upload == nil && hookConfig.Download == nil && hookConfig.Logs == nil && hookConfig.Wait == nil {
			return errors.Errorf("hooks[%d].command, hooks[%d].logs, hooks[%d].wait, hooks[%d].download or hooks[%d].upload is required", index, index, index, index, index)
		}
		enabled := 0
		if hookConfig.Command != "" {
			enabled++
		}
		if hookConfig.Download != nil {
			enabled++
		}
		if hookConfig.Upload != nil {
			enabled++
		}
		if hookConfig.Logs != nil {
			enabled++
		}
		if hookConfig.Wait != nil {
			enabled++
		}
		if enabled > 1 {
			return errors.Errorf("you can only use one of hooks[%d].command, hooks[%d].logs, hooks[%d].wait, hooks[%d].upload and hooks[%d].download per hook", index, index, index, index, index)
		}
		if hookConfig.Upload != nil && hookConfig.Container == nil {
			return errors.Errorf("hooks[%d].container is required if hooks[%d].upload is used", index, index)
		}
		if hookConfig.Download != nil && hookConfig.Container == nil {
			return errors.Errorf("hooks[%d].container is required if hooks[%d].download is used", index, index)
		}
		if hookConfig.Logs != nil && hookConfig.Container == nil {
			return errors.Errorf("hooks[%d].container is required if hooks[%d].logs is used", index, index)
		}
		if hookConfig.Wait != nil && hookConfig.Container == nil {
			return errors.Errorf("hooks[%d].container is required if hooks[%d].wait is used", index, index)
		}
		if hookConfig.Wait != nil && !hookConfig.Wait.Running && hookConfig.Wait.TerminatedWithCode == nil {
			return errors.Errorf("hooks[%d].wait.running or hooks[%d].wait.terminatedWithCode is required if hooks[%d].wait is used", index, index, index)
		}
		if hookConfig.Container != nil {
			if hookConfig.Container.ContainerName != "" && len(hookConfig.Container.LabelSelector) == 0 {
				return errors.Errorf("hooks[%d].container.containerName is defined but hooks[%d].container.labelSelector is not defined", index, index)
			}
			if len(hookConfig.Container.LabelSelector) == 0 && hookConfig.Container.ImageSelector == "" {
				return errors.Errorf("hooks[%d].container.labelSelector and hooks[%d].container.imageSelector are not defined", index, index)
			}
		}
	}

	return nil
}

func validateDeployments(config *latest.Config) error {
	for index, deployConfig := range config.Deployments {
		if deployConfig.Name == "" {
			return errors.Errorf("deployments[%s].name is required", index)
		}
		if encoding.IsUnsafeName(deployConfig.Name) {
			return fmt.Errorf("deployments[%s].name %s has to match the following regex: %v", index, deployConfig.Name, encoding.UnsafeNameRegEx.String())
		}
		if deployConfig.Helm == nil && deployConfig.Kubectl == nil {
			return errors.Errorf("Please specify either helm or kubectl as deployment type in deployment %s", deployConfig.Name)
		}
		if deployConfig.Kubectl != nil && deployConfig.Kubectl.Manifests == nil {
			return errors.Errorf("deployments[%s].kubectl.manifests is required", index)
		}
	}

	return nil
}

func ValidateComponentConfig(deployConfig *latest.DeploymentConfig, overwriteValues map[string]interface{}) error {
	if deployConfig.Helm != nil && deployConfig.Helm.Chart == nil {
		b, err := yaml.Marshal(overwriteValues)
		if err != nil {
			return errors.Errorf("deployments[%s].helm: Error marshaling overwrite values: %v", deployConfig.Name, err)
		}

		componentValues := &latest.ComponentConfig{}
		decoder := yaml.NewDecoder(bytes.NewReader(b))
		decoder.KnownFields(true)
		err = decoder.Decode(componentValues)
		if err != nil {
			return errors.Errorf("deployments[%s].helm.componentChart: component values are incorrect: %v", deployConfig.Name, err)
		}
	}

	return nil
}

func validatePullSecrets(config *latest.Config) error {
	for k, ps := range config.PullSecrets {
		if ps.Name == "" {
			return errors.Errorf("pull secret keys cannot be an empty string")
		}
		if encoding.IsUnsafeName(ps.Name) {
			return fmt.Errorf("pullSecrets[%s] has to match the following regex: %v", ps.Name, encoding.UnsafeNameRegEx.String())
		}
		if ps.Registry == "" {
			return errors.Errorf("pullSecrets[%s].registry: cannot be empty", k)
		}
	}

	return nil
}

func validateImages(config *latest.Config) error {
	// images lists all the image names in order to check for duplicates
	images := map[string]bool{}
	for imageConfigName, imageConf := range config.Images {
		if imageConfigName == "" {
			return errors.Errorf("images keys cannot be an empty string")
		}
		if encoding.IsUnsafeName(imageConfigName) {
			return fmt.Errorf("images[%s] has to match the following regex: %v", imageConfigName, encoding.UnsafeNameRegEx.String())
		}
		if imageConf == nil {
			return errors.Errorf("images.%s is empty and should at least contain an image name", imageConfigName)
		}
		if imageConf.Image == "" {
			return errors.Errorf("images.%s.image is required", imageConfigName)
		}
		if _, tag, _ := imageselector.GetStrippedDockerImageName(imageConf.Image); tag != "" {
			return errors.Errorf("images.%s.image '%s' can not have tag '%s'", imageConfigName, imageConf.Image, tag)
		}
		if imageConf.Build != nil && imageConf.Build.Custom != nil && imageConf.Build.Custom.Command == "" && len(imageConf.Build.Custom.Commands) == 0 {
			return errors.Errorf("images.%s.build.custom.command or images.%s.build.custom.commands is required", imageConfigName, imageConfigName)
		}
		if images[imageConf.Image] {
			return errors.Errorf("multiple image definitions with the same image name are not allowed")
		}
		if imageConf.RebuildStrategy != "" && imageConf.RebuildStrategy != latest.RebuildStrategyDefault && imageConf.RebuildStrategy != latest.RebuildStrategyAlways && imageConf.RebuildStrategy != latest.RebuildStrategyIgnoreContextChanges {
			return errors.Errorf("images.%s.rebuildStrategy %s is invalid. Please choose one of %v", imageConfigName, string(imageConf.RebuildStrategy), []latest.RebuildStrategy{latest.RebuildStrategyAlways, latest.RebuildStrategyIgnoreContextChanges})
		}
		if imageConf.Build != nil && imageConf.Build.Kaniko != nil && imageConf.Build.Kaniko.EnvFrom != nil {
			for _, v := range imageConf.Build.Kaniko.EnvFrom {
				o, err := yaml.Marshal(v)
				if err != nil {
					return errors.Errorf("images.%s.build.kaniko.envFrom is invalid: %v", imageConfigName, err)
				}

				err = jsonyaml.Unmarshal(o, &k8sv1.EnvVarSource{})
				if err != nil {
					return errors.Errorf("images.%s.build.kaniko.envFrom is invalid: %v", imageConfigName, err)
				}
			}
		}
		images[imageConf.Image] = true
	}

	return nil
}

func validateDev(config *latest.Config) error {
	for devPodName, devPod := range config.Dev {
		if devPodName == "" {
			return errors.Errorf("dev[%s] is required", devPodName)
		}
		if encoding.IsUnsafeName(devPodName) {
			return fmt.Errorf("dev[%s] has to match the following regex: %v", devPodName, encoding.UnsafeNameRegEx.String())
		}
		if len(devPod.LabelSelector) == 0 && devPod.ImageSelector == "" {
			return errors.Errorf("dev[%s]: image selector and label selector are nil", devPodName)
		}

		definedSelectors := 0
		if devPod.ImageSelector != "" {
			definedSelectors++
		}
		if len(devPod.LabelSelector) > 0 {
			definedSelectors++
		}
		if definedSelectors > 1 {
			return errors.Errorf("dev[%s]: image selector and label selector cannot all be defined", devPodName)
		}

		err := validateDevContainer(fmt.Sprintf("dev[%s]", devPodName), &devPod.DevContainer, false)
		if err != nil {
			return err
		}
		if len(devPod.Containers) > 0 {
			for i, c := range devPod.Containers {
				err := validateDevContainer(fmt.Sprintf("dev[%s].containers[%s]", devPodName, i), c, true)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validateDevContainer(path string, devContainer *latest.DevContainer, nameRequired bool) error {
	if nameRequired && devContainer.Container == "" {
		return errors.Errorf("%s.container is required", path)
	}

	if !ValidContainerArch(devContainer.Arch) {
		return errors.Errorf("%s.arch is not valid '%s'", path, devContainer.Arch)
	}
	for index, sync := range devContainer.Sync {
		// Validate initial sync strategy
		if !ValidInitialSyncStrategy(sync.InitialSync) {
			return errors.Errorf("%s.sync[%d].initialSync is not valid '%s'", path, index, sync.InitialSync)
		}
		if sync.OnUpload != nil {
			for j, e := range sync.OnUpload.Exec {
				if e.Command == "" {
					return errors.Errorf("%s.sync[%d].exec[%d].command is required", path, index, j)
				}
			}
		}
	}
	for index, port := range devContainer.PortMappingsReverse {
		if port.LocalPort == nil || *port.LocalPort == 0 {
			return errors.Errorf("%s.forward[%d].port is required", path, index)
		}
	}
	for j, p := range devContainer.PersistPaths {
		if p.Path == "" {
			return errors.Errorf("%s.persistPaths[%d].path is required", path, j)
		}
	}

	return nil
}

func EachDevContainer(devPod *latest.DevPod, each func(devContainer *latest.DevContainer) bool) {
	if len(devPod.Containers) > 0 {
		for _, devContainer := range devPod.Containers {
			cont := each(devContainer)
			if !cont {
				break
			}
		}
	} else {
		_ = each(&devPod.DevContainer)
	}
}
