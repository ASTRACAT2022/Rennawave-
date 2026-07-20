import { createZodDto } from 'nestjs-zod';

import { CreateInternalSquadCommand } from '@libs/contracts/commands';

export class CreateInternalSquadRequestDto extends createZodDto(
    CreateInternalSquadCommand.RequestSchema,
) {}

export class CreateInternalSquadResponseDto extends createZodDto(
    CreateInternalSquadCommand.ResponseSchema,
) {}
