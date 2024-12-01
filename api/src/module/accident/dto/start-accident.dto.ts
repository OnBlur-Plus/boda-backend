import { IsDefined, IsEnum, IsString, IsUUID } from 'class-validator';
import { AccidentType } from '../entities/accident.entity';

export class StartAccidentDto {
  @IsEnum(AccidentType)
  type: AccidentType;

  @IsString()
  reason: string;

  @IsUUID()
  streamKey: string;

  @IsDefined()
  startAt = new Date();
}
