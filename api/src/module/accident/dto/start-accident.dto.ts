import { IsEnum, IsUUID } from 'class-validator';
import { AccidentType } from '../entities/accident.entity';

export class StartAccidentDto {
  @IsEnum(AccidentType)
  type: AccidentType;

  @IsUUID()
  streamKey: string;
}
