/*
 * .-'_.---._'-.
 * ||####|(__)||   Protect your secrets, protect your business.
 *   \\()|##//       Secure your sensitive data with Aegis.
 *    \\ |#//                  <aegis.z2h.dev>
 *     .\_/.
 */

package main

import (
	"context"
	"log"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func dontpanic(s string) {
	return
}

func main() {
	log.Println("Will initialize an mTLS Safe server")

	// TODO: get this from environment.
	const socketPath = "unix:///spire-agent-socket/agent.sock"

	ctx := context.Background()

	source, err := workloadapi.NewX509Source(
		ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(socketPath)),
	)

	if err != nil {
		log.Println("Unable to create X509 source")
	} else {
		svid, err := source.GetX509SVID()
		if err != nil {
			// 2023/01/03 19:37:58 svid.id spiffe://aegis.z2h.dev/ns/default/sa/default/n/aegis-workload-demo-559877fd7d-92rcn
			log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! could not get svid")
		}
		log.Println("svid.id", svid.ID)

		log.Println("Everything is awesome!", source)
	}

	// TODO:
	//
	// 1. Create a server that responds to /v1/fetch
	// 2. Fetch workload SVID and bundle from SPIRE
	// 3. In the handler of /v1/fetch verify the caller, and if it is okay respond.
	// 4. Implement the following endpoints too:
	// 4.1. /v1/secret : from sentinel to me to upsert secret
	// 4.2. /healthz
	// 4.3. /readyz

	// TODO: below needs cleanup
	//
	//	// TODO: convert these to tasks. Also it needs a lot of cleanup since
	//	// we are going full steam SPIRE.
	//	dontpanic(`
	//addon:
	//it looks like leveraging SPIRE and go Spiffe
	//will simplify a lot of the hassle:
	//1. dispatching identities to workloads and safe and notary automatically.
	//2. damn easy mtls over grpc between them
	//3. no need for a notary token. svids replace that
	//4. admin token is still needed but can be swapped with an IDM/OIDC
	//5. grpc is faster, lighter
	//6. etc.
	//
	//see https://github.com/spiffe/spire-controller-manager/blob/main/demo/config/greeter-client-id.yaml
	//for a k8s nativr spire id assignment to a workload.
	//
	//To create secrets of any kind, the administrator will require the admin token.
	//
	//To securely deliver the admin token to the administrator:
	//
	//1.
	//Safe will have an API to generate a private/public key pay.
	//Via that API, both the private key and the public key will be
	//delivered to the admin to store them safely outside source control.
	//Private key in a safe medium; public key can be public.
	//
	//This API will deliver the keypair only once. A second request will
	//result in an empty json response or some form of a 4xx error.
	//
	//2.
	//Safe will store the public key and the private key in memory.
	//
	//3.
	//The admin can ask for the admin token, and Safe will return an encrypted
	//(with the public key) version of master key to the admin.
	//
	//4.
	//The admin will use a tool (that exists outside the cluster) to decrypt
	//the encrypted admin token to get their plain text admin token.
	//
	//The admin then will need to store the (unencrypted) admin token safely too.
	//
	//5.
	//Safe will store the encrypted versions of admin token and notary token on
	//a persistent volume.
	//
	//6.
	//Safe will periodically encrypt and store the secrets on that persistent
	//volume too.
	//
	//7.
	//During (1) Safe will send the private key to Notary too.
	//Notary will then ''update'' a Kubernetes secret (that only it can access)
	//that stores the private key.
	//Securing that secret is important, and cluster administrators will
	//be expected to provide proper RBAC for that. Only Notary shall be allowed
	//to read and write to that secret.
	//
	//8. When notary crashes, it still has access to the secret and thus to the
	//private key so it can ask the encrypted notary token from Safe and decrypt it
	//and store it in memory and life is back to normal.
	//
	//9. When Safe crashes, Notary can bootstrap it again because it already has
	//the decrypted notary token in memory.
	//
	//Note: right now, the bootstrap flow is a bit funky.
	//Notary should check the annotation of Safe and bootstrap it while checking
	//for the created pods. This way, whenever Safe is crashed and recreated
	//notary can rebootstrap it whenever Notary is ready.
	//
	//If needed, Notary "can" register all workloads with new workload id and workload
	//secrets. (note: Notary will not need to do this since Safe will likely have
	//the workload id and secrets encrypted on the PV too)
	//
	//9.2. Safe has crashed, but Safe still has the secrets in its PV in encrypted form.
	//It needs the private key to decrypt them. During this (after-crash) bootstrapping,
	//Safe can ask Notary its private key because it knows that it has encrypted data
	//on its drive but it does not have any private key to decrypt it.
	//Notary can deliver the key that it knows to Safe. And Safe then can decrypt what’s
	//on disk and store it in memory. And Safe is back to business too.
	//Until the data is restored back in memory, Safe will not respond to queries
	//from the workloads.
	//
	//10. Since these interactions are in cluster they are moderately secure from
	//a practical standpoint. However, at phase 2, we'd better ensure that at
	//least the private key and token exchanges are done through mTLS.
	//
	//The x.509 certificates (for mTLS) can be autogenerated during deployment and
	//stored as Secrets that are mounted to Safe and Notary.
	//
	//Addenda:
	//
	//A1. Since we are talking about two workloads only, rolling out a PKI with something
	//like SPIFFE will be an overkill. However, if SPIFFE/SPIRE already exists on
	//the system, we can configure Aegis to leverage that too (instead of generating
	//their own x.509 certs).
	//
	//A2. Key and certificate rotation is the top topic of a follow-up issue.
	//Since Notary is the central orchestrator, it can coordinate key rotation too.
	//For example, it can send the new notary token to Safe and all workloads
	//and safe and workloads can cache past 2-3 notary tokens to be on the safe side.
	//
	//Safe can automatically rotate the admin token and tell the admin that their
	//token has expired; then admin will request their new encrypted token and decrypt
	//it with their private key as usual.
	//
	//Notary can periodically rotate workload secrets and notify Safe about them too.
	//
	//However, keeping things simpler initially, and ensuring that one thing works
	//well before adding additional orchestration is "very" important.
	//
	//A3. Any API that the Admin executes can be triggered via CI (like Jenkins) too.
	//So this thing is fully automatable, and can be integrated with a GitOps pipeline
	//etc.
	//`)
	//	apiV1 := &v1Network.Api{}
	//
	//	r := mux.NewRouter()
	//
	//	// Bind handlers.
	//	v1Network.Init(apiV1, v1Service.NewApiV1Service())
	//
	//	// Bind other routes.
	//	route.Probes(r)
	//	route.AdminEndpoints(r, apiV1)
	//	route.WorkloadEndpoints(r, apiV1)
	//	route.NotaryEndpoints(r, apiV1)
	//
	//	p, a := env.Port(), env.AppName()
	//	log.Printf("[SAFE]: '%s' will listen at port '%s'.", a, p)
	//	log.Fatal(http.ListenAndServe(p, r))

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				log.Println("Tick")
			}
		}
	}()
	select {}
}
