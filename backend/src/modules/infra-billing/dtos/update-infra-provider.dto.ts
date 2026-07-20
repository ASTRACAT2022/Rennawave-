import { createZodDto } from 'nestjs-zod';

import { UpdateInfraProviderCommand } from '@libs/contracts/commands';

export class UpdateInfraProviderRequestDto extends createZodDto(
    UpdateInfraProviderCommand.RequestSchema,
) {}

export class UpdateInfraProviderResponseDto extends createZodDto(
    UpdateInfraProviderCommand.ResponseSchema,
) {}
