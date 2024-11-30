import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Accident } from './entities/accident.entities';
import { AccidentController } from './accident.controller';
import { AccidentService } from './accident.service';
import { StreamModule } from 'src/module/stream/stream.module';
import { StreamService } from 'src/module/stream/stream.service';

@Module({
  imports: [TypeOrmModule.forFeature([Accident]), StreamModule],
  controllers: [AccidentController],
  providers: [AccidentService, StreamService],
})
export class AccidentModule {}
