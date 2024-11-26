import { Controller, Get, Param, ParseIntPipe, Query } from '@nestjs/common';
import { AccidentService } from './accident.service';

@Controller('accident')
export class AccidentController {
  constructor(private readonly accidentService: AccidentService) {}

  @Get()
  findAllRecentAccidents(@Query('limit', ParseIntPipe) limit: number = 5) {}

  @Get('count')
  getCountOfAccidents(@Query('date') date: string) {}

  @Get('stream/:streamKey')
  getAccidentsByStreamKey(@Param('streamKey') streamKey: string) {}

  @Get('detail/:id')
  getAccident(@Param('id', ParseIntPipe) id: number) {}
}
