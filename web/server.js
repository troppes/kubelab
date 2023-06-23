import { handler } from './handler.js';
import express from 'express';
import fs from 'fs';
import http from 'http';
import https from 'https';

function sleep(ms) {
    return new Promise((resolve) => {
        setTimeout(resolve, ms);
    });
}


let credentials;
let done = false;
while (!done) {
    try {
        let privateKey = fs.readFileSync(process.env.PRIVATE_KEY, 'utf8');
        let certificate = fs.readFileSync(process.env.CERTIFICATE, 'utf8');
        credentials = { key: privateKey, cert: certificate };
        done = true;
    } catch (e) {
        await sleep(10000)
        console.error(e);
    }
}

const app = express();

const httpServer = http.createServer(app);
const httpsServer = https.createServer(credentials, app);

const PORT = process.env.PORT || 80;
const SSLPORT = process.env.SSLPORT || 443;

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