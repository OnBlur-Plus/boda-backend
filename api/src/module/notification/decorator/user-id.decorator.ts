import { createParamDecorator, ExecutionContext } from '@nestjs/common';

export const UserID = createParamDecorator((_, context: ExecutionContext) => {
  return context.switchToHttp().getRequest().userId;
});
