## Building

To create a production version of your app:

```bash
npm run build
```

This process outputs the App in the `release` folder. The custom server is added to the `release` folder under the name server.js. The app can be locally server by using `npm server`. !Warning! This functions expects the .env files to not contain any comments.


## Certificates

The App expects, when deployed in Docker a mount on the path `/data`, that contains the following files: `cert.pem` and `key.pem`. These files are used to serve the app over HTTPS. If you want to use a different path, you can change it in `src/server.js`. As well as the ca.crt for the connection to the Kubernetes cluster.

