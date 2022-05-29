package db_test

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-appstudio/managed-gitops/backend-shared/config/db"
)

var _ = Describe("RepositoryCredentials Tests", func() {
	var (
		err                  error
		ctx                  context.Context
		clusterUser          *db.ClusterUser
		clusterCredentials   db.ClusterCredentials
		gitopsEngineCluster  db.GitopsEngineCluster
		gitopsEngineInstance db.GitopsEngineInstance
		dbq                  db.AllDatabaseQueries
	)

	When("When ClusterUser, ClusterCredentials, GitopsEngine and GitopsInstance exist", func() {
		BeforeEach(func() {
			// Connect to the database (the connection closes at AfterEach)
			err = db.SetupForTestingDBGinkgo()
			Expect(err).To(BeNil())

			dbq, err = db.NewUnsafePostgresDBQueries(true, true)
			Expect(err).To(BeNil())

			ctx = context.Background()

			// Satisfy the foreign key constraint 'fk_clusteruser_id'
			// aka: ClusterUser.Clusteruser_id) required for 'repo_cred_user_id'
			clusterUser = &db.ClusterUser{
				Clusteruser_id: "fake-repocred-user-id",
				User_name:      "fake-repocred-user",
			}
			err = dbq.CreateClusterUser(ctx, clusterUser)
			Expect(err).To(BeNil())

			// Satisfy the foreign key constraint 'fk_gitopsengineinstance_id'
			// aka: GitOpsEngineInstance.Gitopsengineinstance_id) required for 'repo_cred_gitopsengineinstance_id'
			clusterCredentials = db.ClusterCredentials{
				Clustercredentials_cred_id:  "fake-repocred-Clustercredentials_cred_id",
				Host:                        "fake-repocred-host",
				Kube_config:                 "fake-repocred-kube-config",
				Kube_config_context:         "fake-repocred-kube-config-context",
				Serviceaccount_bearer_token: "fake-repocred-serviceaccount_bearer_token",
				Serviceaccount_ns:           "fake-repocred-Serviceaccount_ns",
			}

			err = dbq.CreateClusterCredentials(ctx, &clusterCredentials)
			Expect(err).To(BeNil())

			gitopsEngineCluster = db.GitopsEngineCluster{
				Gitopsenginecluster_id: "fake-repocred-Gitopsenginecluster_id",
				Clustercredentials_id:  clusterCredentials.Clustercredentials_cred_id,
			}

			err = dbq.CreateGitopsEngineCluster(ctx, &gitopsEngineCluster)
			Expect(err).To(BeNil())

			gitopsEngineInstance = db.GitopsEngineInstance{
				Gitopsengineinstance_id: "fake-repocred-Gitopsengineinstance_id",
				Namespace_name:          "fake-repocred-Namespace_name",
				Namespace_uid:           "fake-repocred-Namespace_uid",
				EngineCluster_id:        gitopsEngineCluster.Gitopsenginecluster_id,
			}

			err = dbq.CreateGitopsEngineInstance(ctx, &gitopsEngineInstance)
			Expect(err).To(BeNil())
		})
		AfterEach(func() {
			// Delete Cluster User, Cluster Credentials, Gitops Engine Cluster, Gitops Engine Instance.
			_, err = dbq.DeleteGitopsEngineInstanceById(ctx, gitopsEngineInstance.Gitopsengineinstance_id)
			Expect(err).To(BeNil())

			_, err = dbq.DeleteGitopsEngineClusterById(ctx, gitopsEngineCluster.Gitopsenginecluster_id)
			Expect(err).To(BeNil())

			_, err = dbq.DeleteClusterCredentialsById(ctx, clusterCredentials.Clustercredentials_cred_id)
			Expect(err).To(BeNil())

			_, err = dbq.DeleteClusterUserById(ctx, clusterUser.Clusteruser_id)
			Expect(err).To(BeNil())

			dbq.CloseDatabase() // Close the database connection.
		})

		It("it should create and delete RepositoryCredentials", func() {

			// Create a RepositoryCredentials object.
			gitopsRepositoryCredentials := db.RepositoryCredentials{
				PrimaryKeyID:    "fake-repo-cred-id",
				UserID:          clusterUser.Clusteruser_id, // constrain 'fk_clusteruser_id'
				PrivateURL:      "https://fake-private-url",
				AuthUsername:    "fake-auth-username",
				AuthPassword:    "fake-auth-password",
				AuthSSHKey:      "fake-auth-ssh-key",
				SecretObj:       "fake-secret-obj",
				EngineClusterID: gitopsEngineInstance.Gitopsengineinstance_id, // constrain 'fk_gitopsengineinstance_id'
				SeqID:           0,
			}

			// Insert the RepositoryCredentials to the database.
			err = dbq.CreateRepositoryCredentials(ctx, &gitopsRepositoryCredentials)
			Expect(err).To(BeNil())

			// Get the RepositoryCredentials from the database.
			fetch, err := dbq.GetRepositoryCredentialsByID(ctx, gitopsRepositoryCredentials.PrimaryKeyID)
			Expect(err).To(BeNil())
			Expect(fetch).Should(Equal(gitopsRepositoryCredentials))

			// Delete the RepositoryCredentials from the database.
			err = dbq.DeleteRepositoryCredentialsByID(ctx, gitopsRepositoryCredentials.PrimaryKeyID)
			Expect(err).To(BeNil())

		})
	})
})
