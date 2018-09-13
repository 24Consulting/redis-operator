package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	redisfailoverv1alpha2 "github.com/spotahome/redis-operator/api/redisfailover/v1alpha2"
	"github.com/spotahome/redis-operator/log"
	mK8SService "github.com/spotahome/redis-operator/mocks/service/k8s"
	rfservice "github.com/spotahome/redis-operator/operator/redisfailover/service"
)

func TestRedisStatefulSetStorageGeneration(t *testing.T) {
	configMapName := rfservice.GetRedisName(generateRF())
	shutdownConfigMapName := rfservice.GetRedisShutdownConfigMapName(generateRF())
	executeMode := int32(0744)
	tests := []struct {
		name           string
		ownerRefs      []metav1.OwnerReference
		expectedSS     appsv1beta2.StatefulSet
		rfRedisStorage redisfailoverv1alpha2.RedisStorage
	}{
		{
			name: "Default values",
			expectedSS: appsv1beta2.StatefulSet{
				Spec: appsv1beta2.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "redis-config",
											MountPath: "/redis",
										},
										{
											Name:      "redis-shutdown-config",
											MountPath: "/redis-shutdown",
										},
										{
											Name:      "redis-data",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "redis-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: configMapName,
											},
										},
									},
								},
								{
									Name: "redis-shutdown-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: shutdownConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
								{
									Name: "redis-data",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{},
									},
								},
							},
						},
					},
				},
			},
			rfRedisStorage: redisfailoverv1alpha2.RedisStorage{},
		},
		{
			name: "Defined an emptydir with storage on memory",
			expectedSS: appsv1beta2.StatefulSet{
				Spec: appsv1beta2.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "redis-config",
											MountPath: "/redis",
										},
										{
											Name:      "redis-shutdown-config",
											MountPath: "/redis-shutdown",
										},
										{
											Name:      "redis-data",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "redis-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: configMapName,
											},
										},
									},
								},
								{
									Name: "redis-shutdown-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: shutdownConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
								{
									Name: "redis-data",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{
											Medium: corev1.StorageMediumMemory,
										},
									},
								},
							},
						},
					},
				},
			},
			rfRedisStorage: redisfailoverv1alpha2.RedisStorage{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		},
		{
			name: "Defined an persistentvolumeclaim",
			expectedSS: appsv1beta2.StatefulSet{
				Spec: appsv1beta2.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "redis-config",
											MountPath: "/redis",
										},
										{
											Name:      "redis-shutdown-config",
											MountPath: "/redis-shutdown",
										},
										{
											Name:      "pvc-data",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "redis-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: configMapName,
											},
										},
									},
								},
								{
									Name: "redis-shutdown-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: shutdownConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
							},
						},
					},
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pvc-data",
							},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									"ReadWriteOnce",
								},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
			},
			rfRedisStorage: redisfailoverv1alpha2.RedisStorage{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							"ReadWriteOnce",
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
		{
			name: "Defined an persistentvolumeclaim with ownerRefs",
			ownerRefs: []metav1.OwnerReference{
				{
					Name: "testing",
				},
			},
			expectedSS: appsv1beta2.StatefulSet{
				Spec: appsv1beta2.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "redis-config",
											MountPath: "/redis",
										},
										{
											Name:      "redis-shutdown-config",
											MountPath: "/redis-shutdown",
										},
										{
											Name:      "pvc-data",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "redis-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: configMapName,
											},
										},
									},
								},
								{
									Name: "redis-shutdown-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: shutdownConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
							},
						},
					},
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pvc-data",
								OwnerReferences: []metav1.OwnerReference{
									{
										Name: "testing",
									},
								},
							},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									"ReadWriteOnce",
								},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
			},
			rfRedisStorage: redisfailoverv1alpha2.RedisStorage{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							"ReadWriteOnce",
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
		{
			name: "Defined an persistentvolumeclaim with ownerRefs keeping the pvc",
			ownerRefs: []metav1.OwnerReference{
				{
					Name: "testing",
				},
			},
			expectedSS: appsv1beta2.StatefulSet{
				Spec: appsv1beta2.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "redis-config",
											MountPath: "/redis",
										},
										{
											Name:      "redis-shutdown-config",
											MountPath: "/redis-shutdown",
										},
										{
											Name:      "pvc-data",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "redis-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: configMapName,
											},
										},
									},
								},
								{
									Name: "redis-shutdown-config",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: shutdownConfigMapName,
											},
											DefaultMode: &executeMode,
										},
									},
								},
							},
						},
					},
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pvc-data",
							},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									"ReadWriteOnce",
								},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
			},
			rfRedisStorage: redisfailoverv1alpha2.RedisStorage{
				KeepAfterDeletion: true,
				PersistentVolumeClaim: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							"ReadWriteOnce",
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		// Generate a default RedisFailover and attaching the required storage
		rf := generateRF()
		rf.Spec.Redis.Storage = test.rfRedisStorage

		generatedStatefulSet := appsv1beta2.StatefulSet{}

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", namespace, mock.Anything).Once().Return(nil, nil)
		ms.On("CreateOrUpdateStatefulSet", namespace, mock.Anything).Once().Run(func(args mock.Arguments) {
			ss := args.Get(1).(*appsv1beta2.StatefulSet)
			generatedStatefulSet = *ss
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)
		err := client.EnsureRedisStatefulset(rf, nil, test.ownerRefs)

		// Check that the storage-related fields are as spected
		assert.Equal(test.expectedSS.Spec.Template.Spec.Volumes, generatedStatefulSet.Spec.Template.Spec.Volumes)
		assert.Equal(test.expectedSS.Spec.Template.Spec.Containers[0].VolumeMounts, generatedStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts)
		assert.Equal(test.expectedSS.Spec.VolumeClaimTemplates, generatedStatefulSet.Spec.VolumeClaimTemplates)
		assert.NoError(err)
	}
}

func TestRedisNameMaximumLength(t *testing.T) {
	tests := []struct {
		name                      string
		givenName                 string
		expectedRedisName         string
		expectedRedisShutdownName string
	}{
		{
			name:                      "Normal Name",
			givenName:                 "normal-name",
			expectedRedisName:         "rfr-normal-name",
			expectedRedisShutdownName: "rfr-shutdown-normal-name",
		},
		{
			name:                      "Long Name",
			givenName:                 "this-is-a-very-long-name-for-an-object-inside-kubernetes-api",
			expectedRedisName:         "rfr-this-is-a-very-long-name-for-an-object-inside-kubernetes",
			expectedRedisShutdownName: "rfr-shutdown-this-is-a-very-long-name-for-an-object-inside-k",
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		// Generate a default RedisFailover and attaching the required storage
		rf := generateRF()
		rf.Name = test.givenName
		rf.Spec.Redis.Exporter = true

		oref := []metav1.OwnerReference{}

		var (
			createdRedisStatefulsetName       string
			createdRedisPDBName               string
			createdRedisServiceName           string
			createdRedisConfigMapName         string
			createdRedisShutdownConfigMapName string
		)

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			a := args.Get(1).(*policyv1beta1.PodDisruptionBudget)
			createdRedisPDBName = a.Name
		}).Return(nil, nil)
		ms.On("CreateOrUpdateStatefulSet", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			a := args.Get(1).(*appsv1beta2.StatefulSet)
			createdRedisStatefulsetName = a.Name
		}).Return(nil)
		ms.On("CreateIfNotExistsService", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			a := args.Get(1).(*corev1.Service)
			createdRedisServiceName = a.Name
		}).Return(nil)
		ms.On("CreateOrUpdateConfigMap", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			a := args.Get(1).(*corev1.ConfigMap)
			createdRedisConfigMapName = a.Name
		}).Return(nil)
		ms.On("CreateOrUpdateConfigMap", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			a := args.Get(1).(*corev1.ConfigMap)
			createdRedisShutdownConfigMapName = a.Name
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)

		assert.NoError(client.EnsureRedisStatefulset(rf, nil, oref))
		assert.Equal(test.expectedRedisName, createdRedisStatefulsetName)
		assert.Equal(test.expectedRedisName, createdRedisPDBName)

		assert.NoError(client.EnsureRedisService(rf, nil, oref))
		assert.Equal(test.expectedRedisName, createdRedisServiceName)

		assert.NoError(client.EnsureRedisConfigMap(rf, nil, oref))
		assert.Equal(test.expectedRedisName, createdRedisConfigMapName)

		assert.NoError(client.EnsureRedisShutdownConfigMap(rf, nil, oref))
		assert.Equal(test.expectedRedisShutdownName, createdRedisShutdownConfigMapName)
	}
}

func TestSentinelNameMaximumLength(t *testing.T) {
	tests := []struct {
		name                 string
		givenName            string
		expectedSentinelName string
	}{
		{
			name:                 "Normal Name",
			givenName:            "normal-name",
			expectedSentinelName: "rfs-normal-name",
		},
		{
			name:                 "Long Name",
			givenName:            "this-is-a-very-long-name-for-an-object-inside-kubernetes-api",
			expectedSentinelName: "rfs-this-is-a-very-long-name-for-an-object-inside-kubernetes",
		},
	}

	for _, test := range tests {
		assert := assert.New(t)

		// Generate a default SentinelFailover and attaching the required storage
		rf := generateRF()
		rf.Name = test.givenName

		oref := []metav1.OwnerReference{}

		var (
			createdSentinelStatefulsetName string
			createdSentinelPDBName         string
			createdSentinelServiceName     string
			createdSentinelConfigMapName   string
		)

		ms := &mK8SService.Services{}
		ms.On("CreateOrUpdatePodDisruptionBudget", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			a := args.Get(1).(*policyv1beta1.PodDisruptionBudget)
			createdSentinelPDBName = a.Name
		}).Return(nil, nil)
		ms.On("CreateOrUpdateDeployment", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			a := args.Get(1).(*appsv1beta2.Deployment)
			createdSentinelStatefulsetName = a.Name
		}).Return(nil)
		ms.On("CreateIfNotExistsService", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			a := args.Get(1).(*corev1.Service)
			createdSentinelServiceName = a.Name
		}).Return(nil)
		ms.On("CreateOrUpdateConfigMap", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
			a := args.Get(1).(*corev1.ConfigMap)
			createdSentinelConfigMapName = a.Name
		}).Return(nil)

		client := rfservice.NewRedisFailoverKubeClient(ms, log.Dummy)

		assert.NoError(client.EnsureSentinelDeployment(rf, nil, oref))
		assert.Equal(test.expectedSentinelName, createdSentinelStatefulsetName)
		assert.Equal(test.expectedSentinelName, createdSentinelPDBName)

		assert.NoError(client.EnsureSentinelService(rf, nil, oref))
		assert.Equal(test.expectedSentinelName, createdSentinelServiceName)

		assert.NoError(client.EnsureSentinelConfigMap(rf, nil, oref))
		assert.Equal(test.expectedSentinelName, createdSentinelConfigMapName)
	}
}