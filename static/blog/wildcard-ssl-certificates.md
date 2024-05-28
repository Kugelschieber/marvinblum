**Published on 7. July 2020**

> _This post was originally published on Medium a while ago. I just added it to my blog for completeness._

![https://api.emvi.com/api/v1/content/NrnkLU0381vIVXCWKMJJ.png](https://marvinblum.de/static/blog/0odQzebaLO/NrnkLU0381vIVXCWKMJJ.png)

Image by [CommitStrip](http://www.commitstrip.com/en/2016/06/13/the-end-of-an-expensive-era/).

When building web applications on top of Kubernetes from a certain point on you will want to make them accessible to the public. Having an SSL certificate is crucial for today's web applications to ensure traffic send between your cluster and your clients is encrypted.

In this article, I will show you how to receive a [wildcard SSL certificate](https://en.wikipedia.org/wiki/Wildcard_certificate) from [Let’s Encrypt](https://letsencrypt.org/) using a DNS01 challenge, cert-manager, and the [ACME DNS server](https://github.com/joohoi/acme-dns).

I will assume you have a cluster up and running with [Helm](https://helm.sh/) installed, exposing some sort of website or service through Ingress and basic knowledge about how to configure Kubernetes objects and Docker. This guide follows a top to bottom approach, which means we will start setting up Ingress first and add everything required to successfully obtain a certificate. Make sure you read completely through it before you start implementing your own solution.

What is a DNS01 challenge and ACME?
-----------------------------------

DNS01 is a certificate authority (CA) challenge method to prove that you are the owner of a specific domain. While using an HTTP01 challenge only proves that you are the owner of a part of a domain (the top-level domain or a subdomain), a DNS01 challenge looks up your DNS configuration to prove that you control the whole domain. This allows CAs, like Let’s Encrypt, to issue wildcard certificates that are valid for all subdomains and your top-level domain. This is especially useful if subdomains are dynamically created like for team or project names in SaaS projects.

ACME is the Automatic Certificate Management Environment protocol and does exactly what it says. It significantly simplifies the process of obtaining SSL certificates and was specifically designed for Let’s Encrypt. It exchanges JSON objects through HTTPS to validate a certificate request.

Kubernetes Ingress SSL certificate setup
----------------------------------------

We start simply by instructing Ingress to consume a secret that contains the certificate we will provide later on. To achieve that, we need to modify its YAML:

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: my-ingress
  annotations:
    kubernetes.io/ingress.class: "nginx"
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
spec:
  tls:
  - hosts:
      - example.com
    secretName: ingress-certificate-secret
  rules:
    # ...
```

By adding the _tls_ section we specified a secret to use for _example.com_. The secret does not exist yet, but will contain our certificates private and public key once it was obtained by cert-manager. The _ssl-redirect_ and _force-ssl-redirect_ annotations instruct Ingress to enforce SSL encryption on all requests. When clients try to reach your service or website through HTTP they will be redirected to HTTPS.

Setting up cert-manager
-----------------------

cert-manager is an add-on to automate the management of SSL certificates inside a Kubernetes cluster. It obtains certificates using various sources and ensures they stay valid. When a certificate gets close to its end of life, cert-manager will try to renew it. This is important since Let’s Encrypt certificates are only valid for up to three months.

To get started with cert-manager we install it first. The easiest way to do that is through Helm (version 0.6 in this example, make sure you install the latest stable version):

```
kubectl apply -f https://raw.githubusercontent.com/jetstack/cert-manager/release-0.6/deploy/manifests/00-crds.yaml
helm install --name cert-manager --version v0.6.0 stable/cert-manager
```

The first command creates a few new Kubernetes custom resources, including _Issuer_ and _Certificate_, which we will configure in a moment.

The second command installs cert-manager and spins up a new pod responsible for issuing certificates and renewing them if required. You can install it for a specific namespace only by providing the namespace argument to Helm.

Next, we create a _Certificate_ object:

```
apiVersion: certmanager.k8s.io/v1alpha1
kind: Certificate
metadata:
  name: my-certificate
spec:
  secretName: ingress-certificate-secret
  issuerRef:
    name: issuer-letsencrypt
  commonName: example.com
  dnsNames:
  - example.com
  - "*.example.com"
  acme:
    config:
    - dns01:
        provider: acmedns
      domains:
        - example.com
        - "*.example.com"
```

The kind is _Certificate_, one of cert-managers custom types we applied before. _secretName_ tells cert-manager the name of the secret to create when the certificate was successfully obtained. This must be identical to the secret name we configured for Ingress. We do not need to create it manually. _issuerRef_ references the _Issuer_ used to get the certificate which we will set up next. _commonName_ represents the server name protected by our SSL certificate and must be equal to the domain or else browsers will complain about it. _dnsNames_ is a list of domains the certificate is valid for. Since we are trying to get a wildcard certificate, the second entry contains an asterisk to mark it valid for all subdomains. The part in the _acme_ section tells cert-manager which challenge type and provider to use. cert-manager supports HTTP01 and DNS01 challenge types as well as a [bunch of different providers](https://cert-manager.readthedocs.io/en/latest/tasks/acme/index.html), including ACME DNS which we use in this article.

To set up the issuer talking to our ACME DNS server we create another object of kind _Issuer_:

```
apiVersion: certmanager.k8s.io/v1alpha1
kind: Issuer
metadata:
  name: issuer-letsencrypt
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: yourname@example.com
    privateKeySecretRef:
      name: account-private-key-secret
    dns01:
      providers:
      - name: acmedns
        acmedns:
          host: https://acme.example.com
          accountSecretRef:
            name: certmanager-secret
            key: acmedns.json
```

The acme section tells cert-manager that this is an ACME issuer. The server, email and privateKeySecretRef are used when the issuer contacts Let’s Encrypt to issue a certificate. If you like to check if that works before you get a production certificate, change the server URL to `https://acme-staging-v02.api.letsencrypt.org/directory` or else you might run into Let’s Encrypts certificate quota. The privateKeySecretRef will be created by cert-manager and stores your Let’s Encrypt private account key. The email meight be used by Let’s Encrypt to contact you or send warnings when your certificate is about to expire.

The _dns01_ section describes which ACME server the issuer will contact. This is our private ACME DNS server we will set up next. Set _host_ to the domain pointing to your ACME DNS server. I recommend configuring a custom subdomain in your DNS settings, but you can use a static IP address as well. The _accountSecretRef_ is a Kubernetes secret and must be created manually. It must contain a JSON file with account information for the ACME server. In our case it looks like this:

```
{
  "example.com": {
    "username": "<username>",
    "password": "<password>",
    "fulldomain": "<generated>.acme.example.com",
    "subdomain": "<generated>",
    "allowfrom": []
  },
  "*.example.com": {
    "username": "<username>",
    "password": "<password>",
    "fulldomain": "<generated>.acme.example.com",
    "subdomain": "<generated>",
    "allowfrom": []
  }
}
```

As you can see it contains the domains we have specified, a username and password as well as two generated values for each entry. Usually, ACME DNS is an interactive system where you sign up and receive the CNAME information you need to create in your DNS settings via its REST API. Since cert-manager handles that for us, it needs to know the account information in advance. The required data will be provided by the ACME DNS server after we have set it up and registered an account. Once we have done that, you can put the JSON file into a new secret called _certmanager-secret_ with _acmedns.json_ as its key.

Setting up the ACME DNS server
------------------------------

We use Docker and Compose to set up our ACME DNS server. For this purpose, create a new virtual machine outside of the cluster and install Docker and Compose. It is necessary to allow incoming DNS requests on port 53, which would collide with Kubernete's own DNS service inside the cluster. Make sure you add a firewall rule to allow incoming traffic to reach your server too. Depending on your VM installation, port 53 might be blocked by systemd-resolved. In that case, change the listening interface in the [server configuration](https://github.com/joohoi/acme-dns#configuration) or disable port 53 for systemd-resolved (add line _DNSStubListener=no_ to _/etc/systemd/resolved.conf_). If you like to use Postgres instead of sqlite3, set it up on your VM or elsewhere or use an existing installation and make sure you can connect to it from your VM (I use a managed solution on Google Cloud).

To start ACME DNS through Compose, create a docker-compose.yml somewhere on your VM:

```
version: '3'
services:
  acmedns:
    restart: always
    image: joohoi/acme-dns
    ports:
      - "443:443"
      - "127.0.0.1:53:53"
      - "127.0.0.1:53:53/udp"
      - "80:80"
    volumes:
      - /acmedns:/etc/acme-dns:ro
```

This pulls the latest ACME DNS image and starts listening on port 443/80 for incoming request to the API, as well as port 53 for incoming TXT DNS requests required to fulfil DNS01 challenges. Note that we added a volume in read only mode to configure ACME DNS. Copy the configuration template from the [GitHub repository](https://github.com/joohoi/acme-dns) and edit it to fit your needs. It must be placed inside the _/acmedns_ directory on your server.

Here are some changes I made for the installation:

```
# listen to incoming traffic instead of localhost
listen = "0.0.0.0:53"
# the (sub-)domain we use for our server as well as the zone name
domain = "acme.example.com"
nsname = "acme.example.com"
# A record pointing to your VM as well as an
# NS record specifying that the server is responsible for all subdomains (e.g. foobar.acme.example.com)
records = [
    "acme.exmaple.com. A <static IP of your VM>",
    "acme.exmaple.com. NS acme.example.com.",
]
# listen on port 443 for encrypted API requests
port = "443"
# obtain a certificate from Let's Encrypt
tls = "letsencrypt"
```

Setting _tls_ to _letsencrypt_ ensures our ACME DNS server issues its own SSL certificate for the REST API so that cert-manager does not obtain certificates from an unsecure source.

Now that we have our server up and running, we create an account and get the information required for the _acmedns.json_ we have created earlier (use curl, Postman, … you name it):

```
POST acme.example.com/register
```

The response looks something like this:

```
{
    "allowfrom": [],
    "fulldomain": "8e5700ea-a4bf-41c7-8a77-e990661dcc6a.acme.example.com",
    "password": "htB9mR9DYgcu9bX_afHF62erXaH2TS7bg9KW3F7Z",
    "subdomain": "8e5700ea-a4bf-41c7-8a77-e990661dcc6a",
    "username": "c36f50e8-4632-44f0-83fe-e070fef28a10"
}
```

Extract the values required to fill out the blanks we have left in the _acmedns.json_ and create the secret for it.

Configuring your DNS
--------------------

To allow Let’s Encrypt to validate DNS01 challenges against our ACME DNS server we have to make some changes to our DNS configuration. For the ACME DNS server we create two new entries:

```
A auth <server-ip>
NS auth.example.com auth.example.com
```

This ensures DNS name resolution requests are handled by our ACME DNS server and marks it responsible for all subdomains of _auth.example.com_. When Let’s Encrypt tries to validate that we are the owner of _example.com_, it will look up a TXT record for that domain. To make sure it will look for it on our ACME DNS server rather than our primary DNS server we create a CNAME record pointing to it:

```
CNAME _acme-challenge 312ecaf7-3ae5-40f3-8559-393e73659a96.auth.example.com.
```

Testing
-------

You should now have a nice green lock in your browser bar when visiting your website or service.

In case you don’t, check the logs of your ACME DNS server, that it serves the DNS TXT record (use one of the online tools) and the status of the _Ingress_, _Issuer_ and _Certificate_ Kubernetes objects we have created. You can use the Docker and Kubernetes log command and inspect the created objects. For example, this is what the _Certificate_ and _Issuer_ objects should look like:

```
# check the certificate got issued and is not expired
kubectl describe certificate my-certificate
# output:
Status:
  Conditions:
    Last Transition Time:  2019-03-07T15:18:29Z
    Message:               Certificate is up to date and has not expired
    Reason:                Ready
    Status:                True
    Type:                  Ready
# check the issuer can connect to ACME DNS and the account was registered
kubectl describe issuer issuer-letsencrypt
# output:
Status:
  Acme:
    Uri:  https://acme-v02.api.letsencrypt.org/acme/acct/50103972
  Conditions:
    Last Transition Time:  2019-01-26T13:32:59Z
    Message:               The ACME account was registered with the ACME server
    Reason:                ACMEAccountRegistered
    Status:                True
    Type:                  Ready
```
