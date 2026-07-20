import { createZodDto } from 'nestjs-zod';

import { CreateInfraProviderCommand } from '@libs/contracts/commands';

export class CreateInfraProviderRequestDto extends createZodDto(
    CreateInfraProviderCommand.RequestSchema,
) {}

export class CreateInfraProviderResponseDto extends createZodDto(
    CreateInfraProviderCommand.ResponseSchema,
) {}
