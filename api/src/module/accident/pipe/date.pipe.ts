import { PipeTransform, Injectable, BadRequestException } from '@nestjs/common';
import { format } from 'date-fns';

const DATE_REGEX = /^\d{4}-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])$/;

@Injectable()
export class DatePipe implements PipeTransform<string | undefined, string> {
  transform(value: string | undefined) {
    if (!value) {
      return format(Date.now(), 'yyyy-MM-dd');
    }

    if (!Date.parse(value) || value.match(DATE_REGEX) === null) {
      throw new BadRequestException('Invalid Date Format');
    }

    return value;
  }
}
