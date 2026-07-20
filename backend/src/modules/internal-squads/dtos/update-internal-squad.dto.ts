import { createZodDto } from 'nestjs-zod';

import { UpdateInternalSquadCommand } from '@libs/contracts/commands';

export class UpdateInternalSquadRequestDto extends createZodDto(
    UpdateInternalSquadCommand.RequestSchema,
) {}

export class UpdateInternalSquadResponseDto extends createZodDto(
    UpdateInternalSquadCommand.ResponseSchema,
) {}
