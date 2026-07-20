import { createZodDto } from 'nestjs-zod';

import { UpdateConfigProfileCommand } from '@libs/contracts/commands';

export class UpdateConfigProfileRequestDto extends createZodDto(
    UpdateConfigProfileCommand.RequestSchema,
) {}

export class UpdateConfigProfileResponseDto extends createZodDto(
    UpdateConfigProfileCommand.ResponseSchema,
) {}
