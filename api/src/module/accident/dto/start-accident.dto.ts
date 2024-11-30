import { IsEnum, IsOptional, IsString, IsUUID } from 'class-validator';
import { AccidentType } from '../entities/accident.entities';

export class StartAccidentDto {
  @IsEnum(AccidentType)
  type: AccidentType;

  @IsString()
  @IsOptional()
  reason: string;

  @IsUUID()
  streamKey: string;
}
