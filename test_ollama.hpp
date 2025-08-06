#pragma once

/**
    @namespace Test
    Namespace containing various utility functions and classes.

    @see MyClass, globalFunction()
  */

  /**
    @class MyClass
    A simple class used for testing purposes.

    @brief Provides a method `myFunction()` and a variable `myVariable`.

    @details The `myFunction()` method does not take any parameters and has no return value. It is intended for testing purposes only.

    @param None
    @return None
    @throws No exceptions are thrown by this function.

    @var myVariable
    A private integer variable used within the MyClass class. Its purpose is not specified in this documentation.
  */
namespace Test {
/**
   * @brief The MyClass class is a custom class within the Test namespace.
   *
   * This class contains a single function, myFunction(), and a variable, myVariable.
   *
   * @see globalFunction()
   *
   * @details
   * The myFunction() member function does not take any parameters and does not return a value. It performs some internal operations specific to this class.
   *
   * The myVariable is an integer variable that can be accessed or modified within the scope of the MyClass instance.
   *
   * @note This class is intended for use within the Test namespace and should not be used directly outside of it.
   */
    class MyClass {
    public:
/**
    @brief Performs specific operation on the instance of MyClass.

    This function does not take any parameters.

    @return Void, as it modifies the state of the MyClass instance.

    @warning This function may have side effects on other parts of the system if not used correctly.

    @see MyClass for more information about the class this function belongs to.
   */
   void Test::MyClass::myFunction()
 */
        void myFunction();
/**
    @brief Private integer variable of MyClass, used for internal calculations.

    This variable is not intended to be accessed or modified directly from outside the class.

    @since 1.0.0
    @warning Modifying this variable may lead to unexpected behavior in the application.
    */
        int myVariable;
    };
    
/**
    @brief Performs a global operation within the Test namespace.

    This function does not take any parameters and does not return any value.

    @warning This function may have side effects on the global state of the application.

    @since 1.0.0
*/
    void globalFunction();
}
