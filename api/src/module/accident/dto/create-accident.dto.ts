import { IsDateString, IsEnum, IsOptional, IsString, IsUUID } from 'class-validator';
import { AccidentLevel, AccidentType } from '../entities/accident.entities';

export class CreateAccidentDto {
  @IsDateString()
  startAt: Date;

  @IsDateString()
  endAt: Date;

  @IsEnum(AccidentType)
  type: AccidentType;

  @IsEnum(AccidentLevel)
  level: AccidentLevel;

  @IsString()
  @IsOptional()
  reason: string;

  @IsUUID()
  streamKey: string;
}
