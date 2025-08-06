#pragma once

/**
    * @brief Namespace containing the core logic and data structures of the Test application.
    */
namespace TestNamespace {
/**
 * @brief TestClass is a class responsible for performing various tests.
 *
 * @namespace TestNamespace
 */
    class TestClass {
    public:
/**
 * @brief Description of the publicMethod function.
 *
 * @param param1 Description of the first parameter.
 * @param param2 Description of the second parameter.
 *
 * @return Description of the return value.
 *
 * @throws Exception1 Description of the first exception.
 * @throws Exception2 Description of the second exception.
 */
        void publicMethod();
/**
 * Public field of TestClass.
 *
 * This field stores the public value of the TestClass instance.
 */
        int publicField;
    private:
/**
 * @brief Private method of TestClass.
 *
 * This method is intended for internal use within the TestClass and should not be called directly from outside the class.
 *
 * @throws std::runtime_error If an error occurs during execution.
 */
        void privateMethod();
        int privateField_;
    };
    
    void globalFunction(int param);
    extern int globalVariable;
}
