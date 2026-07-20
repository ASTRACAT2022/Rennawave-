import { createZodDto } from 'nestjs-zod';

import { CreateInfraBillingHistoryRecordCommand } from '@libs/contracts/commands';

export class CreateInfraBillingHistoryRecordRequestDto extends createZodDto(
    CreateInfraBillingHistoryRecordCommand.RequestSchema,
) {}

export class CreateInfraBillingHistoryRecordResponseDto extends createZodDto(
    CreateInfraBillingHistoryRecordCommand.ResponseSchema,
) {}
