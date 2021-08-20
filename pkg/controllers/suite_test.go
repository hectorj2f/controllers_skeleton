package controllers_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	// +kubebuilder:scaffold:imports

	"github.com/controllers_skeleton/api/v1beta1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg                  *rest.Config
	k8sClient            client.Client
	testEnv              *envtest.Environment
	mgr                  manager.Manager
	mgrStop              chan struct{}
	sampleImageScanBytes []byte
	sampleImageScan      *v1beta1.ImageScan
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
		o.DestWritter = GinkgoWriter
	}))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDInstallOptions: envtest.CRDInstallOptions{
			ErrorIfPathMissing: true,
			Paths: []string{
				filepath.Join("..", "..", "..", "chart", "crds"),
				filepath.Join("..", "..", "..", "test", "testdata"),
			},
		},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	sampleImageScanBytes, err = ioutil.ReadFile(filepath.Join("..", "..", "..", "samples", myYamlFilePath))
	Expect(err).NotTo(HaveOccurred())

	// Prevent the metrics listener being created
	metrics.DefaultBindAddress = "0"

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	Expect(testEnv.Stop()).To(Succeed())

	// Put the DefaultBindAddress back
	metrics.DefaultBindAddress = ":8080"
})

var _ = BeforeEach(func() {
	By("starting manager")
	// This is required by the LicenseSignatureController tests to check if the license expired on each resync
	syncPeriod := 10 * time.Second
	var err error
	disabled := time.Duration(0)
	mgr, err = manager.New(cfg, manager.Options{Scheme: runtime.NewScheme(), SyncPeriod: &syncPeriod, GracefulShutdownTimeout: &disabled})
	Expect(err).NotTo(HaveOccurred())

	Expect(appsv1.AddToScheme(mgr.GetScheme())).To(Succeed())
	Expect(batchv1.AddToScheme(mgr.GetScheme())).To(Succeed())
	Expect(corev1.AddToScheme(mgr.GetScheme())).To(Succeed())
	Expect(rbacv1.AddToScheme(mgr.GetScheme())).To(Succeed())

	Expect(v1beta1.AddToScheme(mgr.GetScheme())).To(Succeed())

	k8sClient, err = client.New(cfg, client.Options{Scheme: mgr.GetScheme()})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	mgrStop = make(chan struct{})

	deserializer := serializer.NewCodecFactory(mgr.GetScheme()).UniversalDeserializer()
	obj, _, err := deserializer.Decode(sampleImageScanBytes, nil, nil)
	Expect(err).NotTo(HaveOccurred())
	Expect(obj).To(BeAssignableToTypeOf(&v1beta1.ImageScan{}))
	sampleImageScan = obj.(*v1beta1.ImageScan)
	sampleImageScan.Namespace = corev1.NamespaceDefault
	mgr.GetScheme().Default(sampleImageScan)
})

var _ = AfterEach(func() {
	By("stopping manager")
	close(mgrStop)
})
