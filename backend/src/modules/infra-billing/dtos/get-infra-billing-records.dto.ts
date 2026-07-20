import { createZodDto } from 'nestjs-zod';

import { GetInfraBillingHistoryRecordsCommand } from '@libs/contracts/commands';

export class GetInfraBillingHistoryRecordsRequestDto extends createZodDto(
    GetInfraBillingHistoryRecordsCommand.RequestQuerySchema,
) {}

export class GetInfraBillingHistoryRecordsResponseDto extends createZodDto(
    GetInfraBillingHistoryRecordsCommand.ResponseSchema,
) {}
