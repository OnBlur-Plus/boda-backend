import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Accident } from './entities/accident.entities';
import { AccidentController } from './accident.controller';
import { AccidentService } from './accident.service';

@Module({
  imports: [TypeOrmModule.forFeature([Accident])],
  controllers: [AccidentController],
  providers: [AccidentService],
})
export class AccidentModule {}
