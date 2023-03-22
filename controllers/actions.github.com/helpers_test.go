package actionsgithubcom

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const defaultGitHubToken = "gh_token"

func startManagers(t ginkgo.GinkgoTInterface, first manager.Manager, others ...manager.Manager) {
	for _, mgr := range append([]manager.Manager{first}, others...) {
		ctx, cancel := context.WithCancel(context.Background())

		g, ctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			return mgr.Start(ctx)
		})

		t.Cleanup(func() {
			cancel()
			require.NoError(t, g.Wait())
		})
	}
}

func createNamespace(t ginkgo.GinkgoTInterface, client client.Client, cfg *rest.Config) (*corev1.Namespace, manager.Manager) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "testns-autoscaling" + RandStringRunes(5)},
	}

	err := client.Create(context.Background(), ns)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := client.Delete(context.Background(), ns)
		require.NoError(t, err)
	})

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Namespace:          ns.Name,
		MetricsBindAddress: "0",
	})
	require.NoError(t, err)

	return ns, mgr
}

func createDefaultSecret(t ginkgo.GinkgoTInterface, client client.Client, namespace string) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "github-config-secret",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"github_token": []byte(defaultGitHubToken),
		},
	}

	err := client.Create(context.Background(), secret)
	require.NoError(t, err)

	return secret
}
