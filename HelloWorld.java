public static Map getSignature(Object param, String secretKey) {
        ObjectMapper objectMapper = new ObjectMapper();
        Map params;
        try {
        String jsonStr = objectMapper.writeValueAsString(param);
        params = objectMapper.readValue(jsonStr, Map.class);
        } catch (Exception e) {
        throw new RuntimeException("生成签名：转换json失败");
        }
        params.remove("itemList");
        params= removeMapEmptyValue(params);
        //对map参数进行排序生成参数
        Set<String> keysSet = params.keySet();
        Object[] keys = keysSet.toArray();
        Arrays.sort(keys);
        StringBuilder temp = new StringBuilder();
        boolean first = true;
        for (Object key : keys) {
        if (first) {
        first = false;
        } else {
        temp.append("&");
        }
        temp.append(key).append("=");
        Object value = params.get(key);
        String valueString = "";
        if (null != value) {
        valueString = String.valueOf(value);
        }
        temp.append(valueString);
        }

        log.info("加密前参数-->"+temp+ secretKey);

        //根据参数生成签名
        String sign = DigestUtils.sha256Hex(temp.toString() + secretKey).toUpperCase();
        log.info("加密后参数sign-->"+sign);
        if (params.get(TIME_STAMP_KEY) == null) {
        params.put(TIME_STAMP_KEY, System.currentTimeMillis());
        }
        params.put(SIGN_KEY, sign);
        return params;
        }


/**
 *
 * @param paramMap
 * @return
 */
public static Map<String,String> removeMapEmptyValue(Map<String,String> paramMap){
        Set<String> set = paramMap.keySet();
        Iterator<String> it = set.iterator();
        List<String> listKey = new ArrayList<String>();
        while (it.hasNext()) {
        String str = it.next();
        if(paramMap.get(str)==null || "".equals(paramMap.get(str))){
        listKey.add(str) ;
        }
        }
        for (String key : listKey) {
        paramMap.remove(key);
        }
        return paramMap;
        }