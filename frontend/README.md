Ziele: Frontend soll innerhalb des Clusters laufen, also SA nutzen

Erstmal nur User Flow:

User soll sich mit token einloggen, token hat die ID und anhand derer alles aus seinen NS ziehen


Danach

isTeacher hat zugriff auf alle deployments mit dem label der klasse: (RBAC geht da nicht, also listen am besten SSR)


# create-svelte

Everything you need to build a Svelte project, powered by [`create-svelte`](https://github.com/sveltejs/kit/tree/master/packages/create-svelte).

## Creating a project

If you're seeing this, you've probably already done this step. Congrats!

```bash
# create a new project in the current directory
npm create svelte@latest

# create a new project in my-app
npm create svelte@latest my-app
```

## Developing

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

```bash
npm run dev

# or start the server and open the app in a new browser tab
npm run dev -- --open
```

## Building

To create a production version of your app:

```bash
npm run build
```

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://kit.svelte.dev/docs/adapters) for your target environment.


## Certificates

The App expects, when deployed in Docker a mount on the path `/data`, that contains the following files: `cert.pem` and `key.pem`. These files are used to serve the app over HTTPS. If you want to use a different path, you can change it in `src/server.js`. As well as the ca.crt for the connection to the Kubernetes cluster.

