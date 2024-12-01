import { Body, Controller, DefaultValuePipe, Delete, Get, Param, ParseIntPipe, Post, Query } from '@nestjs/common';
import { AccidentService } from './accident.service';
import { DatePipe } from './pipe/date.pipe';
import { UpdateAccidentDto } from './dto/update-accident.dto';
import { StartAccidentDto } from './dto/start-accident.dto';
import { EndAccidentDto } from 'src/module/accident/dto/end-accident.dto';
import { IsPublic } from '../auth/auth.guard';

@Controller('accident')
export class AccidentController {
  constructor(private readonly accidentService: AccidentService) {}

  @Get()
  async findAllAccidents(
    @Query('pageNum', new DefaultValuePipe(1), ParseIntPipe) pageNum: number,
    @Query('pageSize', new DefaultValuePipe(10), ParseIntPipe) pageSize: number,
  ) {
    return await this.accidentService.getAccidents(pageNum, pageSize);
  }

  @Get('date')
  async getAccidentsByDate(@Query('date', DatePipe) date: string) {
    return await this.accidentService.getAccidentsByDate(new Date(date));
  }

  @Get('stream/:streamKey')
  async getAccidentsByStreamKey(@Param('streamKey') streamKey: string) {
    return await this.accidentService.getAccidentsByStreamKey(streamKey);
  }

  @Get('detail/:id')
  async getAccident(@Param('id', ParseIntPipe) id: number) {
    return await this.accidentService.getAccidentWithStream(id);
  }

  @Post()
  @IsPublic()
  async startAccident(@Body() startAccidentDto: StartAccidentDto) {
    return await this.accidentService.startAccident(startAccidentDto);
  }

  @Post('end')
  @IsPublic()
  async endAccident(@Body() endAccidentDto: EndAccidentDto) {
    return await this.accidentService.endAccident(endAccidentDto);
  }

  @Post(':accidentId')
  async upateAccident(
    @Param('accidentId', ParseIntPipe) accidentId: number,
    @Body() updateAccidentDto: UpdateAccidentDto,
  ) {
    return await this.accidentService.updateAccident(accidentId, updateAccidentDto);
  }

  @Delete(':accidentId')
  async deleteAccident(@Param('accidentId', ParseIntPipe) accidentId: number) {
    return await this.accidentService.deleteAccident(accidentId);
  }
}
