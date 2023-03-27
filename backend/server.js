import express from 'express';
import bodyParser from 'body-parser';
import {authenticated, authMiddleware} from './middleware/auth.middleware.js';
import {hasRights} from "./middleware/permission.middleware.js";
import authRoutes from './routes/auth.routes.js';
import userRoutes from './routes/user.routes.js';
import * as dotenv from 'dotenv';
import cors from 'cors';

// Stop on Ctrl + c in docker
process.on('SIGINT', function() {
    process.exit();
});

dotenv.config()

const port = process.env.PORT;
export const app = express();

app.listen(port, () => console.log(`KubeLab is listening on ${port}!`));
app.use(bodyParser.json());
app.use(cors());
app.use(authMiddleware);

app.use('/api/auth', authRoutes);
app.use('/api/users', authenticated, hasRights, userRoutes);