import { createZodDto } from 'nestjs-zod';

import { GetRawSubscriptionByShortUuidCommand } from '@libs/contracts/commands/subscription';

export class GetRawSubscriptionByShortUuidRequestDto extends createZodDto(
    GetRawSubscriptionByShortUuidCommand.RequestSchema,
) {}
export class GetRawSubscriptionByShortUuidResponseDto extends createZodDto(
    GetRawSubscriptionByShortUuidCommand.ResponseSchema,
) {}

export class GetRawSubscriptionByShortUuidRequestQueryDto extends createZodDto(
    GetRawSubscriptionByShortUuidCommand.RequestQuerySchema,
) {}
