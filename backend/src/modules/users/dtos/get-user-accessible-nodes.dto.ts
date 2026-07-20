import { createZodDto } from 'nestjs-zod';

import { GetUserAccessibleNodesCommand } from '@libs/contracts/commands';

export class GetUserAccessibleNodesRequestDto extends createZodDto(
    GetUserAccessibleNodesCommand.RequestSchema,
) {}
export class GetUserAccessibleNodesResponseDto extends createZodDto(
    GetUserAccessibleNodesCommand.ResponseSchema,
) {}
