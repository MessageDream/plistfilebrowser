//
//  main.m
//  IPAUtils
//
//  Created by xingbin on 14-7-16.
//
//

#import <Foundation/Foundation.h>
#import "NXJsonSerializer.h"
#include <stdio.h>
bool loadBundleInfo(NSString *path)
{
    NSBundle *bundle = [NSBundle bundleWithPath:path];
    NSDictionary *dict = [NSDictionary dictionaryWithDictionary:bundle.infoDictionary];
    NXJsonSerializer* s = [[NXJsonSerializer alloc] init];
    NSString* ret = [s serialize:dict];
   // BOOL succeed = [ret writeToFile:@"info.json"  atomically:YES encoding:NSUTF8StringEncoding error:nil];
    printf("%s",[ret cStringUsingEncoding:NSUTF8StringEncoding]);
//    if (succeed) {
//        NSLog(@"gen json  ok");
//        return true;
//    }else{
//        NSLog(@"gen json  fail");
//        return false;
//    }
    if (ret) {
        return true;
    }else{
        return false;
    }
}

int main(int argc, const char * argv[])
{
    
    @autoreleasepool {
        
        
        const char *commandc = argv[1];
        const char *paramsc = argv[2];
        NSString *command = [NSString stringWithUTF8String:commandc];
        NSString *path = [NSString stringWithUTF8String:paramsc];
        if (command==nil) {
            NSLog(@"miss command param");
            
        }
        if (path==nil) {
            NSLog(@"miss path param");
            
        }
        if ([command isEqualToString:@"-r"]) {
            loadBundleInfo(path);
        }
        
        
    }
}
