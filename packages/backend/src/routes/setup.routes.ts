import { Router, Request, Response, NextFunction } from 'express';
import bcrypt from 'bcryptjs';
import { Prisma } from '../../generated/prisma/index.js';
import { AppError } from '../middleware/error-handler.js';
import { validatePassword } from '../utils/validate-password.js';

export const setupRoutes: Router = Router();

// GET /api/setup/status — Check if initial setup is needed
setupRoutes.get('/status', async (req: Request, res: Response, next) => {
  try {
    const prisma = req.app.locals.prisma;
    const userCount = await prisma.user.count();
    res.json({ needsSetup: userCount === 0 });
  } catch (err) { next(err); }
});

// POST /api/setup/init — Create the first admin user
setupRoutes.post('/init', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const prisma = req.app.locals.prisma;

    // Only allow if no users exist
    const userCount = await prisma.user.count();
    if (userCount > 0) throw new AppError(403, 'Setup already completed');

    const username = typeof req.body?.username === 'string' ? req.body.username.trim() : '';
    const password = typeof req.body?.password === 'string' ? req.body.password : '';
    const displayNameRaw = typeof req.body?.displayName === 'string' ? req.body.displayName.trim() : '';
    if (!username || !password) throw new AppError(400, 'Username and password are required');

    const pwError = validatePassword(password);
    if (pwError) throw new AppError(400, pwError);

    const passwordHash = await bcrypt.hash(password, 12);
    try {
      await prisma.user.create({
        data: {
          username,
          passwordHash,
          displayName: displayNameRaw || username,
          role: 'admin',
        },
      });
    } catch (e) {
      if (e instanceof Prisma.PrismaClientKnownRequestError && e.code === 'P2002') {
        throw new AppError(409, 'Username is already taken');
      }
      throw e;
    }

    res.status(201).json({ success: true, message: 'Admin account created. You can now log in.' });
  } catch (err) { next(err); }
});
