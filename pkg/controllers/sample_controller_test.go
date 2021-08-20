package controllers_test

import (
	"context"

	"github.com/controllers-test-samples/api/v1beta1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("ControllerName", func() {
	var (
		reconciler         *common.StandardizedController
		scanNamespacedName client.ObjectKey
	)

	BeforeEach(func(done Done) {
		handler := controllers.NewMyControllerHandler()
		reconciler = common.NewStandardizedController(
			handler,
			mgr.GetClient(),
			mgr.GetEventRecorderFor("ControllerName"),
			mgr.GetScheme(),
			log.NewLogger(logf.Log),
		)
		Expect(reconciler.SetupWithManager(mgr)).To(Succeed())
		go func() {
			defer GinkgoRecover()
			Expect(mgr.Start(mgrStop)).NotTo(HaveOccurred())
		}()

		scanNamespacedName, _ = client.ObjectKeyFromObject(sampleImageScan)
		// Create the cluster
		Expect(k8sClient.Create(context.Background(), sampleImageScan)).To(Succeed())

		close(done)
	})

	It("should create image scane", func() {
		// Handle deletion logic in a defer block so always cleaned up.
		defer func() {
			Expect(k8sClient.Delete(context.Background(), sampleImageScan)).To(Succeed())
			Within30Seconds(func() error {
				return k8sClient.Get(context.Background(), scanNamespacedName, &v1beta1.ImageScan{})
			}).Should(BeANotFoundAPIError())
		}()

		scanGVK, err := apiutil.GVKForObject(sampleImageScan, mgr.GetScheme())
		Expect(err).NotTo(HaveOccurred())
		ownerRef := *metav1.NewControllerRef(sampleImageScan, schema.GroupVersionKind{Group: scanGVK.Group, Version: scanGVK.Version, Kind: scanGVK.Kind})

		var imageScanjobCluster *batchv1.Job
		Within30Seconds(func() (*batchv1.Job, error) {
			job = &batchv1.Job{}
			err := k8sClient.Get(context.Background(), scanNamespacedName, imageScanjobCluster)
			return job, err
		}).Should(SatisfyAll(
			BeNamed(sampleImageScan.Name),
			BeInNamespace(sampleImageScan.Namespace),
			BeOwnedBy(ownerRef),
		))
		defer func() {
			Expect(k8sClient.Delete(context.Background(), imageScanjobCluster)).To(Succeed())
			Within30Seconds(func() error {
				return k8sClient.Get(context.Background(), scanNamespacedName, &batchv1.Job{})
			}).Should(BeANotFoundAPIError())
		}()

	})
})
