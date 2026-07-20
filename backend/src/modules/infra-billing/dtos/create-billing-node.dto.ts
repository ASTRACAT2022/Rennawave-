import { createZodDto } from 'nestjs-zod';

import { CreateInfraBillingNodeCommand } from '@libs/contracts/commands';

export class CreateInfraBillingNodeRequestDto extends createZodDto(
    CreateInfraBillingNodeCommand.RequestSchema,
) {}

export class CreateInfraBillingNodeResponseDto extends createZodDto(
    CreateInfraBillingNodeCommand.ResponseSchema,
) {}
