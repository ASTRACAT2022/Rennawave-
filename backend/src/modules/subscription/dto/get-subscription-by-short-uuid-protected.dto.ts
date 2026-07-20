import { createZodDto } from 'nestjs-zod';

import { GetSubscriptionByShortUuidProtectedCommand } from '@libs/contracts/commands';

export class GetSubscriptionByShortUuidProtectedRequestDto extends createZodDto(
    GetSubscriptionByShortUuidProtectedCommand.RequestSchema,
) {}

export class GetSubscriptionByShortUuidProtectedResponseDto extends createZodDto(
    GetSubscriptionByShortUuidProtectedCommand.ResponseSchema,
) {}
