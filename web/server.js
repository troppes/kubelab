import { handler } from './handler.js';
import { env } from './env.js';
import express from 'express';
import fs from 'fs';
import http from 'http';
import https from 'https';

const privateKey = fs.readFileSync(env.PRIVATE_KEY, 'utf8');
const certificate = fs.readFileSync(env.CERTIFICATE, 'utf8');
const credentials = { key: privateKey, cert: certificate };

const app = express();

const httpServer = http.createServer(app);
const httpsServer = https.createServer(credentials, app);

const PORT = env.PORT || 80;
const SSLPORT = env.SSLPORT || 443;

httpServer.listen(PORT, function () {
    console.log('HTTP Server is running on: http://localhost:%s', PORT);
});

httpsServer.listen(SSLPORT, function () {
    console.log('HTTPS Server is running on: https://localhost:%s', SSLPORT);
});

// add healthcheck for kubernetes
app.get('/healthcheck', (req, res) => {
    res.end('ok');
});

// let SvelteKit handle everything else, including serving prerendered pages and static assets
app.use(handler);