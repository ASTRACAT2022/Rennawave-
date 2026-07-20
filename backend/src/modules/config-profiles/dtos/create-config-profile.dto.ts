import { createZodDto } from 'nestjs-zod';

import { CreateConfigProfileCommand } from '@libs/contracts/commands';

export class CreateConfigProfileRequestDto extends createZodDto(
    CreateConfigProfileCommand.RequestSchema,
) {}

export class CreateConfigProfileResponseDto extends createZodDto(
    CreateConfigProfileCommand.ResponseSchema,
) {}
