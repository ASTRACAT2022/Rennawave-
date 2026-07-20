import { createZodDto } from 'nestjs-zod';

import { AddUsersToInternalSquadCommand } from '@libs/contracts/commands';

export class AddUsersToInternalSquadRequestDto extends createZodDto(
    AddUsersToInternalSquadCommand.RequestSchema,
) {}

export class AddUsersToInternalSquadResponseDto extends createZodDto(
    AddUsersToInternalSquadCommand.ResponseSchema,
) {}
