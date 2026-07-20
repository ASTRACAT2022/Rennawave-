import { createZodDto } from 'nestjs-zod';

import { GetUserSubscriptionRequestHistoryCommand } from '@libs/contracts/commands';

export class GetUserSubscriptionRequestHistoryRequestDto extends createZodDto(
    GetUserSubscriptionRequestHistoryCommand.RequestSchema,
) {}
export class GetUserSubscriptionRequestHistoryResponseDto extends createZodDto(
    GetUserSubscriptionRequestHistoryCommand.ResponseSchema,
) {}
