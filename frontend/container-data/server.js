import { handler } from './handler.js';
import express from 'express';
import fs from 'fs';
import http from 'http';
import https from 'https';

const privateKey = fs.readFileSync('./key.pem', 'utf8');
const certificate = fs.readFileSync('./cert.pem', 'utf8');
const credentials = { key: privateKey, cert: certificate };

process.env["AUTH_TRUST_HOST"] = "true";

const app = express();

const httpServer = http.createServer(app);
const httpsServer = https.createServer(credentials, app);

const PORT = 80;
const SSLPORT = 443;

httpServer.listen(PORT, function () {
    console.log('HTTP Server is running on: http://localhost:%s', PORT);
});

httpsServer.listen(SSLPORT, function () {
    console.log('HTTPS Server is running on: https://localhost:%s', SSLPORT);
});

// add a route that lives separately from the SvelteKit app
app.get('/healthcheck', (req, res) => {
    res.end('ok');
});

// let SvelteKit handle everything else, including serving prerendered pages and static assets
app.use(handler);