import { z } from 'zod';

export const BaseInternalSquadSchema = z.object({
    uuid: z.string().uuid(),
    name: z.string(),
});
