import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Accident } from './entities/accident.entities';

@Module({
  imports: [TypeOrmModule.forFeature([Accident])],
})
export class AccidentModule {}
