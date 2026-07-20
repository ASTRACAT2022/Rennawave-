import { createZodDto } from 'nestjs-zod';

import { DeleteConfigProfileCommand } from '@libs/contracts/commands';

export class DeleteConfigProfileRequestDto extends createZodDto(
    DeleteConfigProfileCommand.RequestSchema,
) {}
export class DeleteConfigProfileResponseDto extends createZodDto(
    DeleteConfigProfileCommand.ResponseSchema,
) {}
