import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Accident } from './entities/accident.entity';
import { AccidentController } from './accident.controller';
import { AccidentService } from './accident.service';
import { StreamModule } from 'src/module/stream/stream.module';
import { StreamService } from 'src/module/stream/stream.service';
import { NotificationModule } from '../notification/notification.module';

@Module({
  imports: [TypeOrmModule.forFeature([Accident]), StreamModule, NotificationModule],
  controllers: [AccidentController],
  providers: [AccidentService, StreamService],
})
export class AccidentModule {}
