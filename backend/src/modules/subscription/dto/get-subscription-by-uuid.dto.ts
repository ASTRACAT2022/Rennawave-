import { createZodDto } from 'nestjs-zod';

import { GetSubscriptionByUuidCommand } from '@libs/contracts/commands/subscriptions';

export class GetSubscriptionByUuidRequestDto extends createZodDto(
    GetSubscriptionByUuidCommand.RequestSchema,
) {}

export class GetSubscriptionByUuidResponseDto extends createZodDto(
    GetSubscriptionByUuidCommand.ResponseSchema,
) {}
