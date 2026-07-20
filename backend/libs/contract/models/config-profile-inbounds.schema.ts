import { z } from 'zod';

export const ConfigProfileInboundsSchema = z.object({
    uuid: z.string().uuid(),
    profileUuid: z.string().uuid(),
    tag: z.string(),
    type: z.string(),
    network: z.nullable(z.string()),
    security: z.nullable(z.string()),
    port: z.nullable(z.number()),

    rawInbound: z.nullable(z.unknown()),
});
