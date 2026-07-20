import { z } from 'zod';

import { PartialInfraProviderSchema } from './infra-provider.schema';

export const InfraBillingHistoryRecordSchema = z.object({
    uuid: z.string().uuid(),
    providerUuid: z.string().uuid(),
    amount: z.number(),
    billedAt: z
        .string()
        .datetime()
        .transform((str) => new Date(str)),
    provider: PartialInfraProviderSchema.omit({
        createdAt: true,
        updatedAt: true,
        loginUrl: true,
    }),
});
