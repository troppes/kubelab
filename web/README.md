# Kubelab-Web

This folder contains the Svelte Application. To run the server locally please use `npm run dev`. A built version of the app can be deployed with the playbook provided.

## Building

To create a production version of your app:

```bash
npm run build
```

This process outputs the App in the `release` folder. The custom server is added to the `release` folder under the name server.js. The app can be locally server by using `npm server`. !Warning! The function expects the .env files to not contain any comments.


## Certificates

The App expects, when deployed in Docker a mount on the path `/data`, that contains the following files: `cert.pem` and `key.pem`. These files are used to serve the app over HTTPS. If you want to use a different path, you can change it in `src/server.js`. As well as the ca.crt for the connection to the Kubernetes cluster.

